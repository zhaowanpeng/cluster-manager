package crud

import (
	"sync"
	"zhaowanpeng/cluster-manager/internal/types"
)

func AddNodes(nodes []string) error {
	var wg sync.WaitGroup
	resultChan := make(chan types.Result, len(nodes))

	// for _, ip := range nodes {
	// 	wg.Add(1)

	// 	go func(ip string) {
	// 		defer wg.Done()

	// 		node := model.Node{
	// 			IP: ip,
	// 		}

	// 		sshClient, status := utils.SSH_Check(
	// 			ip,
	// 			22,
	// 			"root",
	// 			"password",
	// 			time.Duration(30)*time.Second,
	// 		)

	// 		isConnected := sshClient != nil

	// 		resultChan <- types.Result{IP: ip, Msg: status}
	// 	}(ip)
	// }

	wg.Wait()
	close(resultChan)

	return nil
}
