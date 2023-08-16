package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dns/armdns"
)

const (
	SleepInSeconds  = 300
	DnsTtlInSeconds = 3600
)

func getIpv6() net.IP {
	resp, err := http.Get("https://api64.ipify.org")
	if err != nil {
		_ = fmt.Errorf("Http Err: %s\n", err)
		return nil
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		_ = fmt.Errorf("IO Err: %s\n", err)
		return nil
	}

	ip := net.ParseIP(string(body))
	// reject IPv4 address
	if ip.To4() != nil {
		return nil
	}
	return ip
}

func getIpv4() net.IP {
	resp, err := http.Get("http://ifconfig.me")
	if err != nil {
		_ = fmt.Errorf("Http Err: %s\n", err)
		return nil
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		_ = fmt.Errorf("IO Err: %s\n", err)
		return nil
	}
	ip := net.ParseIP(string(body))
	if ip.To4() == nil {
		return nil
	}
	return ip.To4()
}

func updateDnsRecords(ip net.IP, client *armdns.RecordSetsClient, resouceGroup string, domainName string) {
	ctx := context.TODO()
	ipStr := ip.String()

	switch len(ip) {
	case net.IPv4len:
		resp, err := client.CreateOrUpdate(ctx,
			resouceGroup,
			domainName,
			"@",
			armdns.RecordTypeA,
			armdns.RecordSet{
				Properties: &armdns.RecordSetProperties{
					ARecords: []*armdns.ARecord{
						{
							IPv4Address: &ipStr,
						},
					},
					TTL: to.Ptr(int64(DnsTtlInSeconds)),
				},
			},
			nil,
		)

		if err != nil {
			fmt.Printf("Update DNS V4 failed: %s; %+v\n", err, resp)
		}
	case net.IPv6len:
		resp, err := client.CreateOrUpdate(ctx,
			resouceGroup,
			domainName,
			"@",
			armdns.RecordTypeAAAA,
			armdns.RecordSet{
				Properties: &armdns.RecordSetProperties{
					AaaaRecords: []*armdns.AaaaRecord{
						{
							IPv6Address: &ipStr,
						},
					},
					TTL: to.Ptr(int64(DnsTtlInSeconds)),
				},
			},
			nil,
		)

		if err != nil {
			fmt.Printf("Update DNS V6 failed: %s; %+v\n", err, resp)
		}
	}
}

func main() {
	resouceGroup := os.Getenv("AZURE_RESOURCE_GROUP")
	domainName := os.Getenv("AZURE_DOMAIN_NAME")
	subscriptionID := os.Getenv("AZURE_SUBSCRIPTION_ID")
	// Need to set the following envs:
	// AZURE_TENANT_ID
	// AZURE_CLIENT_ID
	// AZURE_CLIENT_SECRET or AZURE_CLIENT_CERTIFICATE_PATH
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		_ = fmt.Errorf("Failed to get AAD identity: %s\n", err)
	}
	clientFactory, err := armdns.NewClientFactory(subscriptionID,
		cred,
		nil)
	if err != nil {
		_ = fmt.Errorf("Failed to create Client Factory: %s\n", err)
	}
	client := clientFactory.NewRecordSetsClient()

	var previousIpv4 net.IP = nil
	var previousIpv6 net.IP = nil
	for {
		ipv4 := getIpv4()
		ipv6 := getIpv6()
		if !bytes.Equal(ipv4, previousIpv4) {
			fmt.Printf("IPV4 change from %s to %s\n", previousIpv4.String(), ipv4.String())
			updateDnsRecords(ipv4, client, resouceGroup, domainName)
			previousIpv4 = ipv4
		}

		if !bytes.Equal(ipv6, previousIpv6) {
			fmt.Printf("IPV6 change from %s to %s\n", previousIpv6.String(), ipv6.String())
			updateDnsRecords(ipv6, client, resouceGroup, domainName)
			previousIpv6 = ipv6
		}

		time.Sleep(SleepInSeconds * time.Second)
	}
}
