package cmd

import (
	"fmt"
	"os"

	"github.com/lyb88999/PortScan/internal/config"
	"github.com/lyb88999/PortScan/internal/kafka"
	"github.com/lyb88999/PortScan/internal/models"

	"github.com/spf13/cobra"
)

var (
	masscanCmd = &cobra.Command{
		Use:   "masscan",
		Short: "Use masscan to do the port scan, input (ip, port, bandwidth) and return (protocol, ip, port)",
		Run: func(cmd *cobra.Command, args []string) {
			if ip == "" {
				fmt.Println("ip不能为空")
				return
			}
			if port == 0 {
				fmt.Println("port不能为空")
				return
			}
			// bandwidth不输入的话默认值为1000
			fmt.Printf("ip: %s, port: %d, bandwidth: %d\n", ip, port, bandwidth)

			// 处理逻辑：将(ip, port, bandwidth)发送给kafka
			producer, err := kafka.NewSyncProducer([]string{cfg.KafkaHost}, cfg.InTopic)
			if err != nil {
				fmt.Println("failed to new producer: ", err)
				return
			}
			defer func(producer *kafka.Producer) {
				err := producer.Close()
				if err != nil {
					fmt.Println("failed to close the producer: ", err)
				}
				fmt.Println("producer closed")
			}(producer)
			err = producer.Send(models.ScanOptions{IP: ip, Port: port, BandWidth: bandwidth})
			if err != nil {
				fmt.Println("failed to produce msg to kafka: ", err)
				return
			}
			fmt.Println("Sending the scanning task succeeded")
		},
	}
	ip        string
	port      int
	bandwidth int
	cfg       *config.Config
)

func init() {
	rootCmd.AddCommand(masscanCmd)
	masscanCmd.Flags().StringVar(&ip, "ip", "", "ip地址")
	masscanCmd.Flags().IntVar(&port, "port", 0, "端口号")
	masscanCmd.Flags().IntVar(&bandwidth, "bandwidth", 1000, "带宽")
	var err error
	// cfg, err = config.LoadConfig("../..")
	cfg, err = config.LoadConfigFromExecutable()
	if err != nil {
		fmt.Println("failed to load config: ", err)
		os.Exit(-1)
	}
}
