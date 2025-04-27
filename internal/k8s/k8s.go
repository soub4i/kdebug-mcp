package k8s

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/olekukonko/tablewriter"
	"github.ibm.com/soub4i/kdebug-mcp/internal/config"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type KHandler struct {
	Clientset *kubernetes.Clientset
}

func (h *KHandler) ChooseCtx() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		c := request.Params.Arguments["context"].(string)

		cmd, err := config.SwitchContext(c)
		if err != nil {
			return nil, fmt.Errorf("failed to list nodes: %w", err)
		}
		h.Clientset = cmd

		return mcp.NewToolResultText(fmt.Sprintf("KDebug will use %s", c)), nil
	}
}

func (h *KHandler) GetPodLogs() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		namespace := request.Params.Arguments["namespace"].(string)
		podName, ok := request.Params.Arguments["name"].(string)
		if !ok || podName == "" {
			return nil, fmt.Errorf("pod 'name' is required to get logs")
		}

		pod, err := h.Clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to list pods: %w", err)
		}

		containerName, _ := request.Params.Arguments["container"].(string)

		if containerName == "" {
			containerName = pod.Spec.Containers[0].Name
		}

		podLogOptions := &corev1.PodLogOptions{
			Container: containerName,
			Follow:    false, // You can make this configurable if needed
		}

		req := h.Clientset.CoreV1().Pods(namespace).GetLogs(podName, podLogOptions)
		podLogs, err := req.Stream(ctx)
		if err != nil {
			return nil, fmt.Errorf("error in opening stream: %w", err)
		}
		defer podLogs.Close()

		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, podLogs)
		if err != nil {
			return nil, fmt.Errorf("error in copy information from podLogs to buffer: %w", err)
		}
		logOutput := buf.String()

		return mcp.NewToolResultText(logOutput), nil
	}
}

func (h *KHandler) ListNodes() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		nodes, err := h.Clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to list nodes: %w", err)
		}

		tableString := &strings.Builder{}
		table := tablewriter.NewWriter(tableString)
		table.SetHeader([]string{"NAME", "STATUS", "ROLES", "AGE", "VERSION"})

		for _, node := range nodes.Items {
			roles := "none"
			if len(node.Spec.Taints) > 0 {
				roles = "control-plane"
			} else if _, ok := node.Labels["node-role.kubernetes.io/worker"]; ok {
				roles = "worker"
			}
			age := time.Since(node.CreationTimestamp.Time).Round(time.Second).String()
			status := "Ready"
			for _, condition := range node.Status.Conditions {
				if condition.Type == "Ready" && condition.Status != "True" {
					status = string(condition.Reason)
					break
				}
			}
			table.Append([]string{node.Name, status, roles, age, node.Status.NodeInfo.KubeletVersion})
		}
		table.Render()
		return mcp.NewToolResultText(tableString.String()), nil
	}
}

func (h *KHandler) ListPods() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		namespace := request.Params.Arguments["namespace"].(string)
		name, ok := request.Params.Arguments["name"].(string)
		listOptions := metav1.ListOptions{}
		if ok && name != "" {
			listOptions.FieldSelector = fmt.Sprintf("metadata.name=%s", name)
		}

		pods, err := h.Clientset.CoreV1().Pods(namespace).List(ctx, listOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to list pods: %w", err)
		}

		tableString := &strings.Builder{}
		table := tablewriter.NewWriter(tableString)
		table.SetHeader([]string{"NAMESPACE", "NAME", "READY", "STATUS", "RESTARTS", "AGE", "NODE"})

		for _, pod := range pods.Items {
			readyContainers := 0
			for _, condition := range pod.Status.Conditions {
				if condition.Type == "Ready" && condition.Status == "True" {
					readyContainers++
				}
			}
			restarts := 0
			for _, containerStatus := range pod.Status.ContainerStatuses {
				restarts += int(containerStatus.RestartCount)
			}
			age := time.Since(pod.CreationTimestamp.Time).Round(time.Second).String()
			table.Append([]string{pod.Namespace, pod.Name, fmt.Sprintf("%d/%d", readyContainers, len(pod.Spec.Containers)), string(pod.Status.Phase), fmt.Sprintf("%d", restarts), age, pod.Spec.NodeName})
		}
		table.Render()
		return mcp.NewToolResultText(tableString.String()), nil
	}
}

func (h *KHandler) ListServices() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		namespace := request.Params.Arguments["namespace"].(string)
		name, ok := request.Params.Arguments["name"].(string)

		listOptions := metav1.ListOptions{}
		if ok && name != "" {
			listOptions.FieldSelector = fmt.Sprintf("metadata.name=%s", name)
		}

		services, err := h.Clientset.CoreV1().Services(namespace).List(ctx, listOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to list services: %w", err)
		}

		tableString := &strings.Builder{}
		table := tablewriter.NewWriter(tableString)
		table.SetHeader([]string{"NAMESPACE", "NAME", "TYPE", "CLUSTER-IP", "PORTS", "AGE"})

		for _, svc := range services.Items {
			age := time.Since(svc.CreationTimestamp.Time).Round(time.Second).String()
			ports := []string{}
			for _, port := range svc.Spec.Ports {
				ports = append(ports, fmt.Sprintf("%d/%s", port.Port, port.Protocol))
			}
			table.Append([]string{svc.Namespace, svc.Name, string(svc.Spec.Type), svc.Spec.ClusterIP, strings.Join(ports, ","), age})
		}
		table.Render()
		return mcp.NewToolResultText(tableString.String()), nil
	}
}

func (h *KHandler) ListDeployments() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		namespace := request.Params.Arguments["namespace"].(string)
		name, ok := request.Params.Arguments["name"].(string)

		listOptions := metav1.ListOptions{}
		if ok && name != "" {
			listOptions.FieldSelector = fmt.Sprintf("metadata.name=%s", name)
		}

		deployments, err := h.Clientset.AppsV1().Deployments(namespace).List(ctx, listOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to list deployments: %w", err)
		}

		tableString := &strings.Builder{}
		table := tablewriter.NewWriter(tableString)
		table.SetHeader([]string{"NAMESPACE", "NAME", "READY", "UP-TO-DATE", "AVAILABLE", "AGE"})

		for _, deploy := range deployments.Items {
			age := time.Since(deploy.CreationTimestamp.Time).Round(time.Second).String()
			table.Append([]string{deploy.Namespace, deploy.Name, fmt.Sprintf("%d/%d", deploy.Status.ReadyReplicas, deploy.Status.Replicas), fmt.Sprintf("%d", deploy.Status.UpdatedReplicas), fmt.Sprintf("%d", deploy.Status.AvailableReplicas), age})
		}
		table.Render()
		return mcp.NewToolResultText(tableString.String()), nil
	}
}

func (h *KHandler) ListStatefulSets() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		namespace := request.Params.Arguments["namespace"].(string)
		name, ok := request.Params.Arguments["name"].(string)

		listOptions := metav1.ListOptions{}
		if ok && name != "" {
			listOptions.FieldSelector = fmt.Sprintf("metadata.name=%s", name)
		}

		statefulSets, err := h.Clientset.AppsV1().StatefulSets(namespace).List(ctx, listOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to list stateful sets: %w", err)
		}

		tableString := &strings.Builder{}
		table := tablewriter.NewWriter(tableString)
		table.SetHeader([]string{"NAMESPACE", "NAME", "READY", "AGE"})

		for _, sts := range statefulSets.Items {
			age := time.Since(sts.CreationTimestamp.Time).Round(time.Second).String()
			table.Append([]string{sts.Namespace, sts.Name, fmt.Sprintf("%d/%d", sts.Status.ReadyReplicas, *sts.Spec.Replicas), age})
		}
		table.Render()
		return mcp.NewToolResultText(tableString.String()), nil
	}
}

func (h *KHandler) ListReplicaSets() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		namespace := request.Params.Arguments["namespace"].(string)
		name, ok := request.Params.Arguments["name"].(string)
		listOptions := metav1.ListOptions{}
		if ok && name != "" {
			listOptions.FieldSelector = fmt.Sprintf("metadata.name=%s", name)
		}

		replicaSets, err := h.Clientset.AppsV1().ReplicaSets(namespace).List(ctx, listOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to list replica sets: %w", err)
		}

		tableString := &strings.Builder{}
		table := tablewriter.NewWriter(tableString)
		table.SetHeader([]string{"NAMESPACE", "NAME", "DESIRED", "CURRENT", "READY", "AGE"})

		for _, rs := range replicaSets.Items {
			age := time.Since(rs.CreationTimestamp.Time).Round(time.Second).String()
			table.Append([]string{rs.Namespace, rs.Name, fmt.Sprintf("%d", *rs.Spec.Replicas), fmt.Sprintf("%d", rs.Status.Replicas), fmt.Sprintf("%d", rs.Status.ReadyReplicas), age})
		}
		table.Render()
		return mcp.NewToolResultText(tableString.String()), nil
	}
}

func (h *KHandler) ListDaemonSets() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		namespace := request.Params.Arguments["namespace"].(string)
		name, ok := request.Params.Arguments["name"].(string)
		listOptions := metav1.ListOptions{}

		if ok && name != "" {
			listOptions.FieldSelector = fmt.Sprintf("metadata.name=%s", name)
		}

		daemonSets, err := h.Clientset.AppsV1().DaemonSets(namespace).List(ctx, listOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to list daemon sets: %w", err)
		}

		tableString := &strings.Builder{}
		table := tablewriter.NewWriter(tableString)
		table.SetHeader([]string{"NAMESPACE", "NAME", "DESIRED", "CURRENT", "READY", "UP-TO-DATE", "AVAILABLE", "NODE SELECTOR", "AGE"})

		for _, ds := range daemonSets.Items {
			age := time.Since(ds.CreationTimestamp.Time).Round(time.Second).String()
			nodeSelector := ""
			for k, v := range ds.Spec.Selector.MatchLabels {
				nodeSelector += fmt.Sprintf("%s=%s,", k, v)
			}
			nodeSelector = strings.TrimSuffix(nodeSelector, ",")
			table.Append([]string{ds.Namespace, ds.Name, fmt.Sprintf("%d", ds.Status.DesiredNumberScheduled), fmt.Sprintf("%d", ds.Status.CurrentNumberScheduled), fmt.Sprintf("%d", ds.Status.NumberReady), fmt.Sprintf("%d", ds.Status.UpdatedNumberScheduled), fmt.Sprintf("%d", ds.Status.NumberAvailable), nodeSelector, age})
		}
		table.Render()
		return mcp.NewToolResultText(tableString.String()), nil
	}
}

func (h *KHandler) ListEvents() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		namespace := request.Params.Arguments["namespace"].(string)
		name, ok := request.Params.Arguments["name"].(string)
		listOptions := metav1.ListOptions{}

		if ok && name != "" {
			listOptions = metav1.ListOptions{
				FieldSelector: fmt.Sprintf("involvedObject.name=%s", name),
			}
		}

		events, err := h.Clientset.CoreV1().Events(namespace).List(ctx, listOptions)
		if err != nil {
			return nil, fmt.Errorf("failed to list events: %w", err)
		}

		tableString := &strings.Builder{}
		table := tablewriter.NewWriter(tableString)
		table.SetHeader([]string{"NAMESPACE", "TYPE", "REASON", "OBJECT", "AGE", "FROM", "MESSAGE"})

		for _, event := range events.Items {
			age := time.Since(event.CreationTimestamp.Time).Round(time.Second).String()
			object := fmt.Sprintf("%s/%s", strings.ToLower(event.InvolvedObject.Kind), event.InvolvedObject.Name)
			table.Append([]string{event.Namespace, event.Type, event.Reason, object, age, fmt.Sprintf("%s/%s", event.Source.Component, event.Source.Host), event.Message})
		}
		table.Render()
		return mcp.NewToolResultText(tableString.String()), nil
	}
}
