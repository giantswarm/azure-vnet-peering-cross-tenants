package main

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-11-01/network"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"
)

var (
	ctx                       = context.Background()
	tenant1Tenantid           string
	tenant1Subscriptionid     string
	tenant1Clientid           string
	tenant1Clientsecret       string
	tenant1AuxiliaryTenantIDs string
	tenant2Tenantid           string
	tenant2Subscriptionid     string
	tenant2Clientid           string
	tenant2Clientsecret       string
	tenant2AuxiliaryTenantIDs string
	tenant1ResourceGroupName  string
	tenant1VirtualNetworkName string
	tenant2ResourceGroupName  string
	tenant2VirtualNetworkName string
)

func main() {
	parseEnvironmentVariables()

	tenant1VnetClient, err := getVnetClient(tenant1Tenantid, tenant1Subscriptionid, tenant1Clientid, tenant1Clientsecret, strings.Split(tenant1AuxiliaryTenantIDs, ","))
	if err != nil {
		panic(err)
	}
	tenant1PeeringClient, err := getVnetPeeringsClient(tenant1Tenantid, tenant1Subscriptionid, tenant1Clientid, tenant1Clientsecret, strings.Split(tenant1AuxiliaryTenantIDs, ","))
	if err != nil {
		panic(err)
	}

	//tenant2VnetClient, err := getVnetClient(tenant2Tenantid, tenant2Subscriptionid, tenant2Clientid, tenant2Clientsecret, strings.Split(tenant2AuxiliaryTenantIDs, ","))
	//if err != nil {
	//	panic(err)
	//}
	//tenant2PeeringClient, err := getVnetPeeringsClient(tenant2Tenantid, tenant2Subscriptionid, tenant2Clientid, tenant2Clientsecret, strings.Split(tenant2AuxiliaryTenantIDs, ","))
	//if err != nil {
	//	panic(err)
	//}

	log.Printf("Checking if tenant1 virtual network %#q exists in resource group %#q", tenant1VirtualNetworkName, tenant1ResourceGroupName)
	tenant1Vnet, err := tenant1VnetClient.Get(ctx, tenant1ResourceGroupName, tenant1VirtualNetworkName, "")
	if err != nil {
		panic(err)
	}

	log.Printf("Checking if tenant2 virtual network %#q exists in resource group %#q", tenant2VirtualNetworkName, tenant2ResourceGroupName)
	tenant2Vnet, err := tenant1VnetClient.Get(ctx, tenant2ResourceGroupName, tenant2VirtualNetworkName, "")
	if err != nil {
		panic(err)
	}

	log.Printf("Ensuring vnet peering %#q exists on the tenant2 vnet %#q in resource group %#q", tenant1VirtualNetworkName, tenant2VirtualNetworkName, tenant2ResourceGroupName)
	_, err = tenant1PeeringClient.CreateOrUpdate(ctx, tenant2ResourceGroupName, tenant2VirtualNetworkName, tenant1VirtualNetworkName, buildPeering(*tenant1Vnet.ID))
	if err != nil {
		panic(err)
	}

	// Create vnet peering on the control plane side.
	log.Printf("Ensuring vnet peering %#q exists on the tenant1 vnet %#q in resource group %#q", tenant2ResourceGroupName, tenant1VirtualNetworkName, tenant1ResourceGroupName)
	_, err = tenant1PeeringClient.CreateOrUpdate(ctx, tenant1ResourceGroupName, tenant1VirtualNetworkName, tenant2ResourceGroupName, buildPeering(*tenant2Vnet.ID))
	if err != nil {
		panic(err)
	}
}

func parseEnvironmentVariables() {
	var ok bool

	// Tenant1
	tenant1ResourceGroupName, ok = os.LookupEnv("TENANT1_RESOURCE_GROUP")
	if !ok {
		panic("TENANT1_RESOURCE_GROUP must be set in the environment")
	}
	tenant1VirtualNetworkName, ok = os.LookupEnv("TENANT1_VIRTUAL_NETWORK")
	if !ok {
		panic("TENANT1_VIRTUAL_NETWORK must be set in the environment")
	}
	tenant1Clientid, ok = os.LookupEnv("TENANT1_AZURE_CLIENTID")
	if !ok {
		panic("TENANT1_AZURE_CLIENTID must be set in the environment")
	}
	tenant1Clientsecret, ok = os.LookupEnv("TENANT1_AZURE_CLIENTSECRET")
	if !ok {
		panic("TENANT1_AZURE_CLIENTSECRET must be set in the environment")
	}
	tenant1Tenantid, ok = os.LookupEnv("TENANT1_AZURE_TENANTID")
	if !ok {
		panic("TENANT1_AZURE_TENANTID must be set in the environment")
	}
	tenant1Subscriptionid, ok = os.LookupEnv("TENANT1_AZURE_SUBSCRIPTIONID")
	if !ok {
		panic("TENANT1_AZURE_SUBSCRIPTIONID must be set in the environment")
	}
	tenant1AuxiliaryTenantIDs, ok = os.LookupEnv("TENANT1_AZURE_AUX_TENANTIDS")
	if !ok {
		panic("TENANT1_AZURE_AUX_TENANTIDS must be set in the environment")
	}

	// Tenant2
	tenant2ResourceGroupName, ok = os.LookupEnv("TENANT2_RESOURCE_GROUP")
	if !ok {
		panic("TENANT2_RESOURCE_GROUP must be set in the environment")
	}
	tenant2VirtualNetworkName, ok = os.LookupEnv("TENANT2_VIRTUAL_NETWORK")
	if !ok {
		panic("TENANT2_VIRTUAL_NETWORK must be set in the environment")
	}
	//tenant2Clientid, ok = os.LookupEnv("TENANT2_AZURE_CLIENTID")
	//if !ok {
	//	panic("TENANT2_AZURE_CLIENTID must be set in the environment")
	//}
	//tenant2Clientsecret, ok = os.LookupEnv("TENANT2_AZURE_CLIENTSECRET")
	//if !ok {
	//	panic("TENANT2_AZURE_CLIENTSECRET must be set in the environment")
	//}
	tenant2Tenantid, ok = os.LookupEnv("TENANT2_AZURE_TENANTID")
	if !ok {
		panic("TENANT2_AZURE_TENANTID must be set in the environment")
	}
	tenant2Subscriptionid, ok = os.LookupEnv("TENANT2_AZURE_SUBSCRIPTIONID")
	if !ok {
		panic("TENANT2_AZURE_SUBSCRIPTIONID must be set in the environment")
	}
	tenant2AuxiliaryTenantIDs, ok = os.LookupEnv("TENANT2_AZURE_AUX_TENANTIDS")
	if !ok {
		panic("TENANT2_AZURE_AUX_TENANTIDS must be set in the environment")
	}
}

func buildPeering(vnetId string) network.VirtualNetworkPeering {
	peering := network.VirtualNetworkPeering{
		VirtualNetworkPeeringPropertiesFormat: &network.VirtualNetworkPeeringPropertiesFormat{
			AllowVirtualNetworkAccess: to.BoolPtr(true),
			AllowForwardedTraffic:     to.BoolPtr(false),
			AllowGatewayTransit:       to.BoolPtr(false),
			UseRemoteGateways:         to.BoolPtr(false),
			RemoteVirtualNetwork: &network.SubResource{
				ID: &vnetId,
			},
		},
	}

	return peering
}

func getVnetPeeringsClient(tenantid, subscriptionid, clientid, clientsecret string, auxiliarytenantids []string) (network.VirtualNetworkPeeringsClient, error) {
	env, err := azure.EnvironmentFromName("AZUREPUBLICCLOUD")
	if err != nil {
		panic(err)
	}

	oauthConfig, err := adal.NewMultiTenantOAuthConfig(env.ActiveDirectoryEndpoint, tenantid, auxiliarytenantids, adal.OAuthOptions{})
	if err != nil {
		panic(err)
	}
	client := network.NewVirtualNetworkPeeringsClientWithBaseURI(env.ResourceManagerEndpoint, subscriptionid)
	token, err := adal.NewMultiTenantServicePrincipalToken(oauthConfig, clientid, clientsecret, env.ServiceManagementEndpoint)
	if err != nil {
		panic(err)
	}
	client.Authorizer = autorest.NewMultiTenantServicePrincipalTokenAuthorizer(token)
	client.RetryAttempts = 1

	return client, nil
}

func getVnetClient(tenantid, subscriptionid, clientid, clientsecret string, auxiliarytenantids []string) (network.VirtualNetworksClient, error) {
	env, err := azure.EnvironmentFromName("AZUREPUBLICCLOUD")
	if err != nil {
		panic(err)
	}

	oauthConfig, err := adal.NewMultiTenantOAuthConfig(env.ActiveDirectoryEndpoint, tenantid, auxiliarytenantids, adal.OAuthOptions{})
	if err != nil {
		panic(err)
	}
	client := network.NewVirtualNetworksClientWithBaseURI(env.ResourceManagerEndpoint, subscriptionid)
	token, err := adal.NewMultiTenantServicePrincipalToken(oauthConfig, clientid, clientsecret, env.ServiceManagementEndpoint)
	if err != nil {
		panic(err)
	}
	client.Authorizer = autorest.NewMultiTenantServicePrincipalTokenAuthorizer(token)
	client.RetryAttempts = 1

	return client, nil
}
