package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"quic-test/internal"
)

func main() {
	var (
		provider        = flag.String("provider", "aws", "Cloud provider (aws, azure, gcp, digitalocean, linode)")
		region          = flag.String("region", "us-east-1", "Cloud region")
		instanceType    = flag.String("instance-type", "t3.medium", "Instance type")
		minInstances    = flag.Int("min-instances", 1, "Minimum number of instances")
		maxInstances    = flag.Int("max-instances", 5, "Maximum number of instances")
		autoScale       = flag.Bool("auto-scale", true, "Enable auto-scaling")
		loadBalancer    = flag.String("load-balancer", "ALB", "Load balancer type")
		sslEnabled      = flag.Bool("ssl", true, "Enable SSL/TLS")
		domainName      = flag.String("domain", "", "Domain name for the deployment")
		enableMonitoring = flag.Bool("monitoring", true, "Enable monitoring")
		enableFirewall  = flag.Bool("firewall", true, "Enable firewall")
		sshKeyName      = flag.String("ssh-key", "", "SSH key name")
		name            = flag.String("name", "quic-tester", "Deployment name")
		action          = flag.String("action", "deploy", "Action (deploy, scale, stop, status)")
		deploymentID    = flag.String("deployment-id", "", "Deployment ID for scale/stop/status actions")
		targetInstances = flag.Int("target-instances", 0, "Target number of instances for scaling")
	)
	flag.Parse()

	fmt.Println("☁️  QUIC Cloud Deployment")
	fmt.Println("=========================")

	// Create cloud config
	config := internal.CloudConfig{
		Provider:        internal.CloudProvider(*provider),
		Region:          *region,
		InstanceType:    *instanceType,
		MinInstances:    *minInstances,
		MaxInstances:    *maxInstances,
		AutoScale:       *autoScale,
		LoadBalancerType: *loadBalancer,
		SSLEnabled:      *sslEnabled,
		DomainName:      *domainName,
		EnableMonitoring: *enableMonitoring,
		EnableFirewall:  *enableFirewall,
		SSHKeyName:      *sshKeyName,
	}

	// Create deployment manager
	manager := internal.NewCloudDeploymentManager()

	switch *action {
	case "deploy":
		deployToCloud(manager, config, *name)
	case "scale":
		scaleDeployment(manager, *deploymentID, *targetInstances)
	case "stop":
		stopDeployment(manager, *deploymentID)
	case "status":
		getDeploymentStatus(manager, *deploymentID)
	case "list":
		listDeployments(manager)
	default:
		log.Fatalf("Unknown action: %s", *action)
	}
}

func deployToCloud(manager *internal.CloudDeploymentManager, config internal.CloudConfig, name string) {
	fmt.Printf("🚀 Deploying to %s cloud...\n", config.Provider)
	fmt.Printf("  Region: %s\n", config.Region)
	fmt.Printf("  Instance Type: %s\n", config.InstanceType)
	fmt.Printf("  Instances: %d-%d\n", config.MinInstances, config.MaxInstances)
	fmt.Printf("  Auto-scaling: %v\n", config.AutoScale)
	fmt.Printf("  Load Balancer: %s\n", config.LoadBalancerType)
	fmt.Printf("  SSL: %v\n", config.SSLEnabled)
	fmt.Printf("  Monitoring: %v\n", config.EnableMonitoring)

	deployment, err := manager.DeployToCloud(config.Provider, config, name)
	if err != nil {
		log.Fatalf("Failed to deploy: %v", err)
	}

	fmt.Printf("\n✅ Deployment successful!\n")
	fmt.Printf("Deployment ID: %s\n", deployment.ID)
	fmt.Printf("Status: %s\n", deployment.Status)
	fmt.Printf("Created: %s\n", deployment.CreatedAt.Format(time.RFC3339))

	if len(deployment.Instances) > 0 {
		fmt.Printf("\nInstances:\n")
		for _, instance := range deployment.Instances {
			fmt.Printf("  %s: %s (%s) - %s\n", 
				instance.ID, instance.Name, instance.PublicIP, instance.Status)
		}
	}

	if deployment.LoadBalancer != nil {
		fmt.Printf("\n⚖️  Load Balancer:\n")
		fmt.Printf("  ID: %s\n", deployment.LoadBalancer.ID)
		fmt.Printf("  Type: %s\n", deployment.LoadBalancer.Type)
		fmt.Printf("  DNS: %s\n", deployment.LoadBalancer.DNSName)
		fmt.Printf("  Port: %d\n", deployment.LoadBalancer.Port)
		fmt.Printf("  SSL: %v\n", deployment.LoadBalancer.SSLEnabled)
	}

	if deployment.Monitoring != nil {
		fmt.Printf("\nMonitoring:\n")
		fmt.Printf("  Provider: %s\n", deployment.Monitoring.Provider)
		fmt.Printf("  Dashboard: %s\n", deployment.Monitoring.DashboardURL)
		fmt.Printf("  Alerts: %d\n", len(deployment.Monitoring.Alerts))
	}
}

func scaleDeployment(manager *internal.CloudDeploymentManager, deploymentID string, targetInstances int) {
	if deploymentID == "" {
		log.Fatal("Deployment ID is required for scaling")
	}

	deployment, exists := manager.GetDeployment(deploymentID)
	if !exists {
		log.Fatalf("Deployment %s not found", deploymentID)
	}

	fmt.Printf("📈 Scaling deployment %s to %d instances...\n", deploymentID, targetInstances)

	// Get the appropriate cloud manager
	// This is a simplified version - in reality, you'd need to track which manager created the deployment
	// For now, we'll assume AWS
	config := deployment.Config
	cloudManager, err := internal.NewCloudManager(config.Provider, config)
	if err != nil {
		log.Fatalf("Failed to create cloud manager: %v", err)
	}

	if err := cloudManager.Scale(nil, deployment, targetInstances); err != nil {
		log.Fatalf("Failed to scale deployment: %v", err)
	}

	fmt.Printf("✅ Scaling completed: %d instances\n", len(deployment.Instances))
}

func stopDeployment(manager *internal.CloudDeploymentManager, deploymentID string) {
	if deploymentID == "" {
		log.Fatal("Deployment ID is required for stopping")
	}

	deployment, exists := manager.GetDeployment(deploymentID)
	if !exists {
		log.Fatalf("Deployment %s not found", deploymentID)
	}

	fmt.Printf("🛑 Stopping deployment %s...\n", deploymentID)

	// Get the appropriate cloud manager
	config := deployment.Config
	cloudManager, err := internal.NewCloudManager(config.Provider, config)
	if err != nil {
		log.Fatalf("Failed to create cloud manager: %v", err)
	}

	if err := cloudManager.Stop(nil, deployment); err != nil {
		log.Fatalf("Failed to stop deployment: %v", err)
	}

	fmt.Printf("✅ Deployment stopped\n")
}

func getDeploymentStatus(manager *internal.CloudDeploymentManager, deploymentID string) {
	if deploymentID == "" {
		log.Fatal("Deployment ID is required for status check")
	}

	deployment, exists := manager.GetDeployment(deploymentID)
	if !exists {
		log.Fatalf("Deployment %s not found", deploymentID)
	}

	fmt.Printf("Deployment Status: %s\n", deploymentID)
	fmt.Printf("========================\n")
	fmt.Printf("Status: %s\n", deployment.Status)
	fmt.Printf("Provider: %s\n", deployment.Config.Provider)
	fmt.Printf("Region: %s\n", deployment.Config.Region)
	fmt.Printf("Created: %s\n", deployment.CreatedAt.Format(time.RFC3339))
	fmt.Printf("Updated: %s\n", deployment.UpdatedAt.Format(time.RFC3339))

	fmt.Printf("\nInstances (%d):\n", len(deployment.Instances))
	for _, instance := range deployment.Instances {
		fmt.Printf("  %s: %s (%s) - %s\n", 
			instance.ID, instance.Name, instance.PublicIP, instance.Status)
	}

	if deployment.LoadBalancer != nil {
		fmt.Printf("\n⚖️  Load Balancer:\n")
		fmt.Printf("  ID: %s\n", deployment.LoadBalancer.ID)
		fmt.Printf("  Type: %s\n", deployment.LoadBalancer.Type)
		fmt.Printf("  DNS: %s\n", deployment.LoadBalancer.DNSName)
		fmt.Printf("  Health: %s\n", deployment.LoadBalancer.HealthCheck)
	}
}

func listDeployments(manager *internal.CloudDeploymentManager) {
	deployments := manager.ListDeployments()

	fmt.Printf("Cloud Deployments (%d):\n", len(deployments))
	fmt.Printf("==========================\n")

	if len(deployments) == 0 {
		fmt.Println("No deployments found")
		return
	}

	for _, deployment := range deployments {
		fmt.Printf("ID: %s\n", deployment.ID)
		fmt.Printf("Name: %s\n", deployment.Name)
		fmt.Printf("Status: %s\n", deployment.Status)
		fmt.Printf("Provider: %s\n", deployment.Config.Provider)
		fmt.Printf("Instances: %d\n", len(deployment.Instances))
		fmt.Printf("Created: %s\n", deployment.CreatedAt.Format(time.RFC3339))
		fmt.Println("---")
	}
}
