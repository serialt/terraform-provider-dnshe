package dnshe

import (
	"fmt"
	"os"
	"testing"
)

func Test_update_record(t *testing.T) {

	apiKey := os.Getenv("DNSHE_API_KEY")
	apiSecret := os.Getenv("DNSHE_API_SECRET")

	client := NewClient("", apiKey, apiSecret)

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
