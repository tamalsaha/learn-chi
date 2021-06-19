package main

import (
	"bytes"
	"context"
	"github.com/unrolled/render"
	"log"
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.wandrs.dev/binding"
	"go.wandrs.dev/inject"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type User struct {
	Name string
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(binding.Injector(render.New()))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello world"))
	})
	r.Get("/inject", binding.Handler(hello))

	r.With(binding.Inject(createKubeClient), binding.Map(User{
		Name: "John",
	})).Get("/k8s", binding.Handler(k8s))

	log.Println("running server on :3333")
	http.ListenAndServe(":3333", r)
}

func hello(r *http.Request) string {
	return "hello " + r.URL.Query().Get("name")
}

func k8s(kc kubernetes.Interface, nodeclient corev1.NodeInterface, u User) []byte {
	var buf bytes.Buffer
	buf.WriteString("hello " + u.Name)
	buf.WriteRune('\n')

	info, err := kc.Discovery().ServerVersion()
	if err != nil {
		panic(err)
	}
	buf.WriteString("k8s version = " + info.GitVersion)
	buf.WriteRune('\n')

	nodes, err := nodeclient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	buf.WriteString("Nodes: \n")
	for _, n := range nodes.Items {
		buf.WriteString(n.Name)
		buf.WriteRune('\n')
	}
	return buf.Bytes()
}

func createKubeClient(injector inject.Injector) {
	masterURL := ""
	kubeconfigPath := filepath.Join(homedir.HomeDir(), ".kube", "config")

	config, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfigPath)
	if err != nil {
		log.Fatalf("Could not get Kubernetes config: %s", err)
	}

	var client kubernetes.Interface = kubernetes.NewForConfigOrDie(config)
	injector.Map(client)

	var nodeclient corev1.NodeInterface = client.CoreV1().Nodes()
	injector.Map(nodeclient)
}
