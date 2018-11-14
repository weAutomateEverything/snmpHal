package main

import "testing"

func TestGetGroupFromCommunitySuccess(t *testing.T){
	result,err := getGroupFromCommunity("T12345")
	if err != nil {
		t.Errorf("Was not expecting an error. Received %v",err)
	}
	if "12345" != result {
		t.Errorf("Expected 12345, received %v",result)
	}
}
