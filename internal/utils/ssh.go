package utils

import (
	"fmt"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

func Get_SSH_Client() {}

func SSH_Check(ip string, port int, user string, pwd string, timeout time.Duration) (*ssh.Client, string) {
	addr := fmt.Sprintf("%s:%d", ip, port)

	config := &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{ssh.Password(pwd)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //
		Timeout:         timeout,
	}

	//检查端口是否可到达
	conn, err := net.DialTimeout("tcp", addr, timeout)

	if err != nil {
		return nil, err.Error()
	}

	defer conn.Close()

	//ssh
	sshConn, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		return nil, err.Error()
	}
	client := ssh.NewClient(sshConn, chans, reqs)
	defer client.Close()

	// 否则返回默认的成功提示
	return client, "ssh success"
}

func Exec_SSH_Command(ip string, port int, user, pwd, command string, timeout time.Duration) (string, error) {
	addr := fmt.Sprintf("%s:%d", ip, port)

	config := &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{ssh.Password(pwd)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         timeout,
	}

	// 先通过 TCP 建立连接
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	// 建立 SSH 连接
	sshConn, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		return "", err
	}
	client := ssh.NewClient(sshConn, chans, reqs)
	defer client.Close()

	// 创建新的 SSH 会话
	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	// 执行命令，捕获标准输出和错误输出（合并输出）
	output, err := session.CombinedOutput(command)
	return string(output), err
}
