package services

import (
	"fabric-sdk-go/server/sdkprovider"
	"fmt"
	"golang.org/x/net/context"

	pb "fabric-sdk-go/protos"
)

type ChannelService struct {
	provider sdkprovider.SdkProvider
}

func NewChannelService() *ChannelService {
	return &ChannelService{
		provider: GetSdkProvider(),
	}
}

func (c *ChannelService) CreateChannel(ctx context.Context, r *pb.CreateChannelRequest) (*pb.CreateChannelResponse, error) {
	transactionID, err := c.provider.CreateChannel(r.ChannelId)
	if err != nil {
		fmt.Println(err)
	}
	return &pb.CreateChannelResponse{TransactionId: transactionID}, nil
}