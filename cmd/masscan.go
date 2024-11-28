package cmd

import (
	"fmt"
	"github.com/lyb88999/PortScan/internal/config"
	"github.com/lyb88999/PortScan/internal/kafka"
	"github.com/lyb88999/PortScan/internal/models"

	"github.com/spf13/cobra"
)

// masscanCmd represents the masscan command
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
			err = producer.Send(models.Data{IP: ip, Port: port, Bandwidth: bandwidth})
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
	cfg, err = config.LoadConfig("../..")
	if err != nil {
		fmt.Println("failed to load config: ", err)
	}

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// masscanCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// masscanCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
