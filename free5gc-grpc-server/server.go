package main

import (
    "context"
    "fmt"
    "log"
    "net"
    "strings"

    "google.golang.org/grpc"
    pb "github.com/s3wong/nephio-code/nfdeploymentrpc"

    apiv1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/dynamic"
    "k8s.io/client-go/rest"
)

type server struct {
    pb.UnimplementedNFDeploymentRPCServer
    dyClient dynamic.Interface
}

func newClient() (dynamic.Interface, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	dynClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return dynClient, nil
}

func GetInterfaceConfigs(interfaceConfigs []*pb.InterfaceConfig, interfaceName string) []*pb.InterfaceConfig {
	var selectedInterfaceConfigs []*pb.InterfaceConfig

	for _, interfaceConfig := range interfaceConfigs {
		if interfaceConfig.Name == interfaceName {
			selectedInterfaceConfigs = append(selectedInterfaceConfigs, interfaceConfig)
		}
	}

	return selectedInterfaceConfigs
}

func GetFirstInterfaceConfig(interfaceConfigs []*pb.InterfaceConfig, interfaceName string) (*pb.InterfaceConfig, error) {
	for _, interfaceConfig := range interfaceConfigs {
		if interfaceConfig.Name == interfaceName {
			return interfaceConfig, nil
		}
	}

	return nil, fmt.Errorf("Interface %s not found", interfaceName)
}

func GetFirstInterfaceConfigIPv4(interfaceConfigs []*pb.InterfaceConfig, interfaceName string) (string, string, string, error) {
	interfaceConfig, err := GetFirstInterfaceConfig(interfaceConfigs, interfaceName)
	if err != nil {
		return "", "", "", err
	}

    if interfaceConfig.Ipv4 == nil {
        return "", "", "", fmt.Errorf("No IPv4 in interface %s", interfaceName)
    }

	ip, netmask, err := net.ParseCIDR(interfaceConfig.Ipv4.Address)
	if err != nil {
		return "", "", "", err
	}

	return ip.String(), interfaceConfig.Ipv4.Gateway, netmask.String(), nil
}

func (s *server) CreateUpdate(ctx context.Context, in *pb.NFDeployment) (*pb.NFDeploymentResponse, error) {
	namespace := in.Namespace
	instanceName := in.Name

	templateValues := new(configurationTemplateValues)
	n4ip, n4Gateway, n4net, err := GetFirstInterfaceConfigIPv4(in.Ifconfig, "n4")
	if err == nil {
		templateValues.N4ENABLED = true
		templateValues.N4SUBNET = n4ip
		templateValues.N4CIDR = strings.Split(n4net, "/")[1]
		templateValues.N4GATEWAY = n4Gateway
		templateValues.N4EXCLUDEIP = n4Gateway
		// TODO: hardcoded values
		templateValues.N4NETWORKNAME = "n4network"
		templateValues.N4CNINAME = "macvlan"
		templateValues.N4CNIMASTERINTF = "eth0"
	} else {
		fmt.Printf("SKW GetFirstInterfaceConfigIPv4 for n4 returns error %v\n", err)
		templateValues.N4ENABLED = false
	}

	n3ip, n3Gateway, n3net, err := GetFirstInterfaceConfigIPv4(in.Ifconfig, "n3")
	if err == nil {
		templateValues.N3ENABLED = true
		templateValues.N3SUBNET = n3ip
		templateValues.N3CIDR = strings.Split(n3net, "/")[1]
		templateValues.N3GATEWAY = n3Gateway
		templateValues.N3EXCLUDEIP = n3Gateway
		// TODO: hardcoded values
		templateValues.N3NETWORKNAME = "n3network"
		templateValues.N3CNINAME = "macvlan"
		templateValues.N3CNIMASTERINTF = "eth0"
	} else {
		fmt.Printf("SKW GetFirstInterfaceConfigIPv4 for n3 returns error %v\n", err)
		templateValues.N3ENABLED = false
	}

	n6ip, n6Gateway, n6net, err := GetFirstInterfaceConfigIPv4(in.Ifconfig, "n6")
	if err == nil {
		templateValues.N6ENABLED = true
		templateValues.N6SUBNET = n6ip
		templateValues.N6CIDR = strings.Split(n6net, "/")[1]
		templateValues.N6GATEWAY = n6Gateway
		templateValues.N6EXCLUDEIP = n6Gateway
		// TODO: hardcoded values
		templateValues.N6NETWORKNAME = "n6network"
		templateValues.N6CNINAME = "macvlan"
		templateValues.N6CNIMASTERINTF = "eth0"
	} else {
		fmt.Printf("SKW GetFirstInterfaceConfigIPv4 for n6 returns error %v\n", err)
		templateValues.N6ENABLED = false
	}

	if configuration, err := renderConfigurationTemplate(*templateValues); err != nil {
		return nil, err
	} else {
		configMap := &apiv1.ConfigMap{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "ConfigMap",
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      instanceName,
			},
			Data: map[string]string{
				"values.yaml": configuration,
			},
		}
		fmt.Printf("ConfigMap generated is %v\n", configMap)
	}

	return &pb.NFDeploymentResponse{
        Condition: "good",
        Errormsg: "",
    }, nil
}

func main() {
    lis, err := net.Listen("tcp", "50051")
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }
    s := grpc.NewServer()
    svr := &server{}
    svr.dyClient, err = newClient()
    if err != nil {
        log.Fatalf("failed to create dynamic client: %v", err)
    }

    pb.RegisterNFDeploymentRPCServer(s, svr)

    log.Printf("Server listening...")
    if err := s.Serve(lis); err != nil {
        log.Fatalf("failed to serve: %v", err)
    }
}
