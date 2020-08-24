/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and

limitations under the License.
*/
package main

import (
	"encoding/json"
	"flag"
	"k8s.io/klog"
	"os"

	meshv1 "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/api/v1"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/controllers"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = meshv1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	conf := config.Config{}
	flag.StringVar(&metricsAddr, "metrics-addr", ":9443", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", true,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&conf.DockerHub, "istio-docker-hub", "", "istio-operator docker hub")
	flag.StringVar(&conf.IstioOperatorCharts, "istiooperator-charts", "", "istio-operator charts")
	flag.StringVar(&conf.ServerAddress, "apigateway-addr", "", "apigateway address")
	flag.StringVar(&conf.UserToken, "user-token", "", "apigateway usertoken to control k8s cluster")
	flag.Parse()
	conf.ServerAddress = "http://9.143.0.40:31000/tunnels/clusters/BCS-K8S-15091/"
	conf.UserToken = "mCdfmlzonNPiAeWhANX1nj91ouBeQckQ"
	conf.IstioOperatorCharts = "./istio-operator"
	conf.IstioOperatorCr = "./istiooperator-cr.yaml"
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		//LeaderElection:     enableLeaderElection,
		//LeaderElectionID:   "333fb49e.tencent.com",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}
	by,_ := json.Marshal(conf)
	klog.Infof("MeshManager config(%s)", string(by))
	if err = (&controllers.MeshClusterReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("MeshCluster"),
		Scheme: mgr.GetScheme(),
		MeshClusters: make(map[string]*controllers.MeshClusterManager),
		Conf: conf,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "MeshCluster")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	func(){
		setupLog.Info("starting manager")
		if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
			setupLog.Error(err, "problem running manager")
			os.Exit(1)
		}
	}()

	/*// New Service
	service := micro.NewService(
		micro.Name("meshmanager.bkbcs.tencent.com"),
		micro.Version("latest"),
	)

	// Initialise service
	service.Init()

	// Register Handler
	mesh.RegisterMeshHandler(service.Server(), new(handler.Mesh))

	// Register Struct as Subscriber
	micro.RegisterSubscriber("meshmanager.bkbcs.tencent.com", service.Server(), new(subscriber.Mesh))

	// Run service
	if err := service.Run(); err != nil {
		log.Fatal(err)
	}*/
}
