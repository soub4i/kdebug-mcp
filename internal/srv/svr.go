package srv

import (
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.ibm.com/soub4i/kdebug-mcp/internal/config"
	"github.ibm.com/soub4i/kdebug-mcp/internal/k8s"
	"k8s.io/client-go/kubernetes"
)

type SHandler struct {
	Server *server.MCPServer
	K8s    *kubernetes.Clientset
}

func NewSHandler() *SHandler {

	clientset, err := config.BuildConfig()
	if err != nil {
		fmt.Printf("Error creating clientset: %v\n", err)
		os.Exit(1)
	}

	return &SHandler{
		Server: server.NewMCPServer(
			"Kubernetes Debug Tools - @soub4i",
			"0.0.1",
			server.WithResourceCapabilities(true, true),
			server.WithLogging(),
			server.WithRecovery(),
		),
		K8s: clientset,
	}
}

func (s *SHandler) RegisterHandlers() {

	contextTool := mcp.NewTool(
		"context",
		mcp.WithDescription("You need to ask the user first which context to use while debuging."),
		mcp.WithString("context", mcp.Required()),
	)

	nodesTool := mcp.NewTool(
		"nodes",
		mcp.WithDescription("List Kubernetes nodes. Use this tool when you need list of nodes of the cluster, don't have the information internally and no other tools can provide this information. Only output printed to stdout or stderr is returned so ALWAYS use print statements!"),
	)

	podsTool := mcp.NewTool(
		"pods",
		mcp.WithDescription("List Kubernetes pods. Use this tool when you need list of pods in given namespace or to get a specific pod by name, don't have the information internally and no other tools can provide this information. Only output printed to stdout or stderr is returned so ALWAYS use print statements!"),
		mcp.WithString("namespace", mcp.Description("Namespace of the pods"), mcp.DefaultString("default")),
		mcp.WithString("name", mcp.Description("Name of the pod"), mcp.DefaultString("")),
	)

	podLogsTool := mcp.NewTool(
		"podLogs",
		mcp.WithDescription("Get Kubernetes pod logs. Use this tool when you need get a logs from a specific pod by its name, don't have the information internally and no other tools can provide this information. Only output printed to stdout or stderr is returned so ALWAYS use print statements!"),
		mcp.WithString("namespace", mcp.Description("Namespace of the pods"), mcp.DefaultString("default")),
		mcp.WithString("name", mcp.Description("Name of the pod"), mcp.Required()),
	)

	servicesTool := mcp.NewTool(
		"services",
		mcp.WithDescription("List Kubernetes services. Use this tool when you need list of services in given namespace or to get a specific service by name, don't have the information internally and no other tools can provide this information. Only output printed to stdout or stderr is returned so ALWAYS use print statements!"),
		mcp.WithString("namespace", mcp.Description("Namespace of the services"), mcp.DefaultString("default")),
		mcp.WithString("name", mcp.Description("Name of the service"), mcp.DefaultString("")),
	)

	deploymentsTool := mcp.NewTool(
		"deployments",
		mcp.WithDescription("List Kubernetes deployments. Use this tool when you need list of deployments in given namespace or to get a specific deployment by name, don't have the information internally and no other tools can provide this information. Only output printed to stdout or stderr is returned so ALWAYS use print statements!"),
		mcp.WithString("namespace", mcp.Description("Namespace of the deployments"), mcp.DefaultString("default")),
		mcp.WithString("name", mcp.Description("Name of the deployment"), mcp.DefaultString("")),
	)

	statefulSetsTool := mcp.NewTool(
		"statefulsets",
		mcp.WithDescription("List Kubernetes statefulsets. Use this tool when you need list of deployments in given namespace  or to get a specific statefulset by name, don't have the information internally and no other tools can provide this information. Only output printed to stdout or stderr is returned so ALWAYS use print statements!"),
		mcp.WithString("namespace", mcp.Description("Namespace of the stateful sets"), mcp.DefaultString("default")),
		mcp.WithString("name", mcp.Description("Name of the statefulset"), mcp.DefaultString("")),
	)

	replicaSetsTool := mcp.NewTool(
		"replicasets",
		mcp.WithDescription("List Kubernetes replicasets. Use this tool when you need list of replicasets in given namespace  or to get a specific replicaset by name, don't have the information internally and no other tools can provide this information. Only output printed to stdout or stderr is returned so ALWAYS use print statements!"),
		mcp.WithString("namespace", mcp.Description("Namespace of the replica sets"), mcp.DefaultString("default")),
		mcp.WithString("name", mcp.Description("Name of the replicaset"), mcp.DefaultString("")),
	)

	daemonSetsTool := mcp.NewTool(
		"daemonsets",
		mcp.WithDescription("List Kubernetes daemonsets. Use this tool when you need list of daemonsets in given namespace  or to get a specific daemonset by name, don't have the information internally and no other tools can provide this information. Only output printed to stdout or stderr is returned so ALWAYS use print statements!"),
		mcp.WithString("namespace", mcp.Description("Namespace of the daemon sets"), mcp.DefaultString("default")),
		mcp.WithString("name", mcp.Description("Name of the daemonset"), mcp.DefaultString("")),
	)

	eventsTool := mcp.NewTool(
		"events",
		mcp.WithDescription("List Kubernetes events. Use this tool when you need list of events in given namespace or need a specific resource by name, don't have the information internally and no other tools can provide this information. Only output printed to stdout or stderr is returned so ALWAYS use print statements!"),
		mcp.WithString("namespace", mcp.Description("Namespace of the events"), mcp.DefaultString("default")),
		mcp.WithString("name", mcp.Description("Name of the events"), mcp.DefaultString("")),
	)

	k := k8s.KHandler{Clientset: s.K8s}

	// --- Add the tools to the server ---
	s.Server.AddTool(contextTool, k.ChooseCtx())
	s.Server.AddTool(nodesTool, k.ListNodes())
	s.Server.AddTool(podsTool, k.ListPods())
	s.Server.AddTool(podLogsTool, k.GetPodLogs())
	s.Server.AddTool(servicesTool, k.ListServices())
	s.Server.AddTool(deploymentsTool, k.ListDeployments())
	s.Server.AddTool(statefulSetsTool, k.ListStatefulSets())
	s.Server.AddTool(replicaSetsTool, k.ListReplicaSets())
	s.Server.AddTool(daemonSetsTool, k.ListDaemonSets())
	s.Server.AddTool(eventsTool, k.ListEvents())

}
