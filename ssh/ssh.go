package ssh

import (
	"bytes"
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"os"
	"path"
)

type Client struct {
	Host       string
	User       string
	Password   string
	PrivateKey string
	SSHClient  *ssh.Client
	SSHSession *ssh.Session
	SFTPClient *sftp.Client
}

func (this *Client) Connect() error {
	conf := ssh.ClientConfig{
		User:            this.User,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	if this.Password != "" {
		conf.Auth = append(conf.Auth, ssh.Password(this.Password))
	}

	if privateKey, err := getPrivateKey(this.PrivateKey); err == nil {
		conf.Auth = append(conf.Auth, privateKey)
	}

	sshClient, err := ssh.Dial("tcp", this.Host, &conf)
	if err != nil {
		return fmt.Errorf("unable to connect: %v", err)
	}
	this.SSHClient = sshClient

	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		return fmt.Errorf("unable to connect: %v", err)
	}
	this.SFTPClient = sftpClient

	return nil
}

// Close the connection
func (this *Client) Close() {
	if this.SFTPClient != nil {
		_ = this.SFTPClient.Close()
	}
	if this.SSHSession != nil {
		_ = this.SSHSession.Close()
	}
	if this.SSHClient != nil {
		_ = this.SSHClient.Close()
	}
}

// Get the private key for current user
func getPrivateKey(privateKey string) (ssh.AuthMethod, error) {

	key := []byte(privateKey)

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("parse private key failed: %v", err)
	}

	return ssh.PublicKeys(signer), nil
}

func (this *Client) newSession() error {

	session, err := this.SSHClient.NewSession()
	if err != nil {
		return fmt.Errorf("unable to create session: %s", err)
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
		return fmt.Errorf("unable to create pty: %v", err)
	}

	this.SSHSession = session
	return nil
}

func (this *Client) Execute(command string) error {

	fmt.Println(command)

	if err := this.newSession(); err != nil {
		return err
	}

	var stdoutBuf bytes.Buffer
	this.SSHSession.Stdout = &stdoutBuf
	if err := this.SSHSession.Run(command); err != nil {
		return err
	}

	fmt.Println(this.SSHSession.Stdout)

	if this.SSHSession != nil {
		_ = this.SSHSession.Close()
	}
	return nil
}

func UploadFile(client *Client, localFilePath string, remotePath string) error {
	remoteFileName := path.Base(localFilePath)
	fmt.Printf("%s -> %s\n", localFilePath, path.Join(remotePath, remoteFileName))

	srcFile, err := os.Open(localFilePath)
	if err != nil {
		return fmt.Errorf("os.Open error : %s", localFilePath)
	}
	defer func() {
		_ = srcFile.Close()
	}()

	dstFile, err := client.SFTPClient.Create(path.Join(remotePath, remoteFileName))
	if err != nil {
		return fmt.Errorf("sftpClient.Create error : %s", path.Join(remotePath, remoteFileName))
	}
	defer func() {
		_ = dstFile.Close()
	}()

	ff, err := ioutil.ReadAll(srcFile)
	if err != nil {
		return fmt.Errorf("readAll error : %s", localFilePath)
	}
	_, err = dstFile.Write(ff)
	if err != nil {
		return fmt.Errorf("write error : %s", err)
	}
	//fmt.Println(localFilePath)

	return nil
}

func UploadDirectory(client *Client, localPath string, remotePath string) error {
	localFiles, err := ioutil.ReadDir(localPath)
	if err != nil {
		log.Fatal("read dir list fail ", err)
	}

	for _, backupDir := range localFiles {
		localFilePath := path.Join(localPath, backupDir.Name())
		remoteFilePath := path.Join(remotePath, backupDir.Name())
		if backupDir.IsDir() {
			if err := client.SFTPClient.Mkdir(remoteFilePath); err != nil {
				return fmt.Errorf("mkdir error : %s", err)
			}
			if err := UploadDirectory(client, localFilePath, remoteFilePath); err != nil {
				return err
			}
		} else {
			if err := UploadFile(client, localFilePath, remotePath); err != nil {
				return err
			}
		}
	}

	return nil
}
