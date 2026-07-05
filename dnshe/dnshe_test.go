package dnshe

import (
	"fmt"
	"os"
	"testing"
)

var client *Client

func init() {
	apiKey := os.Getenv("DNSHE_API_KEY")
	apiSecret := os.Getenv("DNSHE_API_SECRET")
	client = NewClient("", apiKey, apiSecret)
}

func Test_update_record(t *testing.T) {
	records, _ := client.ListDNSRecords(5004916620)
	fmt.Printf("records: %v\n", records)

	apiReq := UpdateDNSRecordRequest{
		ID:      924766,
		Type:    "A",
		Name:    "github",
		Content: "9.9.9.9",
	}
	resp, _ := client.UpdateDNSRecord(apiReq)
	fmt.Println(resp)

}

func Test_addDomain(t *testing.T) {
	resp, err := client.RegisterSubdomain("krabkkk", "bbroot.com")
	if err != nil {
		fmt.Println("err: ", err)
	}
	fmt.Println("resp: ", resp)
}

func Test_listDomain(t *testing.T) {
	resp, err := client.ListSubdomains(ListSubdomainsParams{})
	if err != nil {
		fmt.Printf("err: %v\n", err)
	}
	fmt.Printf("resp: %v\n", resp)

}
