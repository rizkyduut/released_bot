package releasedbot

import (
	"fmt"
	"github.com/rizkyduut/released_bot/dbadapter"
	"io/ioutil"
	"strings"
)

const (
	chatDefault              = "Perintah tidak dikenal. /help aja dulu."
	chatIncompleteCommand    = "perintah tidak dikenal. `/%s help` aja dulu"
	chatHelpNotFound         = "/%s help belum tersedia."
	chatDeploySuccess        = "Berhasil mengupdate latest release branch %s"
	chatServiceAdded         = "Service %s berhasil ditambahkan."
	chatServiceDeleted       = "Service %s berhasil dihapus."
	chatServiceNotRegistered = "Service %s belum terdaftar sebelumnya."
	chatGroupAdded           = "Group %s berhasil ditambahkan."
	chatGroupNotRegistered   = "Group %s belum terdaftar sebelumnya."
	chatListAll              = "- `%s`\n"
	chatListLatest           = "- `%s` branch `%s` by %s\n"
)

type BotData struct {
	Sender           string
	RawMessage       string
	Command          string
	CommandArguments string
}

type ReleasedBot struct {
	dbAdapter dbadapter.DBAdapter
}

type Handler func(b *BotData) (string, error)

func New(db dbadapter.DBAdapter) *ReleasedBot {
	b := &ReleasedBot{
		dbAdapter: db,
	}

	return b
}

func (rb *ReleasedBot) DefaultHandler(b *BotData) (string, error) {
	return chatDefault, nil
}

func (rb *ReleasedBot) HelpHandler(b *BotData) (string, error) {
	return readHelpFile("help"), nil
}

func (rb *ReleasedBot) LatestHandler(b *BotData) (string, error) {
	if b.CommandArguments == "help" {
		return readHelpFile("latest"), nil
	}

	args := strings.Fields(b.CommandArguments)
	if len(args) == 0 {
		return "", fmt.Errorf(chatIncompleteCommand, "latest")
	}

	group := args[0]
	allService, err := rb.dbAdapter.GetAllServiceInGroup(group)
	if err != nil {
		return "", err
	}

	msg := fmt.Sprintf("Daftar latest release branch di group %s:\n", group)
	for _, service := range allService {
		branchAndModifier, err := rb.dbAdapter.GetService(service)
		if err != nil {
			return "", err
		}

		tmp := strings.Split(branchAndModifier, "|")
		msg = msg + fmt.Sprintf(chatListLatest, service, tmp[0], tmp[1])
	}

	return msg, nil
}

func (rb *ReleasedBot) DeployHandler(b *BotData) (string, error) {
	if b.CommandArguments == "help" {
		return readHelpFile("deploy"), nil
	}

	args := strings.Fields(b.CommandArguments)
	if len(args) < 2 {
		return "", fmt.Errorf(chatIncompleteCommand, "deploy")
	}

	service := args[0]
	_, err := rb.dbAdapter.GetService(service)
	if err != nil {
		// TODO: should be use chatServiceNotRegistered if err == redis.Nil
		return "", err
	}

	branchAndModifier := args[1] + "|" + b.Sender
	err = rb.dbAdapter.DeployService(service, branchAndModifier)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(chatDeploySuccess, service), nil
}

func (rb *ReleasedBot) ServiceHandler(b *BotData) (string, error) {
	if b.CommandArguments == "help" {
		return readHelpFile("service"), nil
	}

	args := strings.Fields(b.CommandArguments)
	if len(args) == 0 {
		return "", fmt.Errorf(chatIncompleteCommand, "service")
	}

	cmd := args[0]
	switch cmd {
	case "add":
		if len(args) < 3 {
			return "", fmt.Errorf(chatIncompleteCommand, "service")
		}

		group := args[1]
		service := args[2]
		allGroup, err := rb.dbAdapter.GetAllGroup()
		if !(contains(allGroup, group)) {
			return fmt.Sprintf(chatGroupNotRegistered, group), nil
		}

		err = rb.dbAdapter.AddGroupService(group, service)
		if err != nil {
			return "", err
		}

		err = rb.dbAdapter.AddService(service)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf(chatServiceAdded, service), nil
	case "delete":
		if len(args) < 3 {
			return "", fmt.Errorf(chatIncompleteCommand, "service")
		}

		group := args[1]
		service := args[2]
		allGroup, err := rb.dbAdapter.GetAllGroup()
		if !(contains(allGroup, group)) {
			return fmt.Sprintf(chatGroupNotRegistered, group), nil
		}

		err = rb.dbAdapter.DeleteGroupService(group, service)
		if err != nil {
			return "", err
		}

		err = rb.dbAdapter.DeleteService(service)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf(chatServiceDeleted, service), nil
	case "group":
		if len(args) < 2 {
			return "", fmt.Errorf(chatIncompleteCommand, "service")
		}

		group := args[1]
		allService, err := rb.dbAdapter.GetAllServiceInGroup(group)
		if err != nil {
			return "", err
		}

		msg := fmt.Sprintf("Daftar semua service di group %s:\n", group)
		for _, service := range allService {
			msg = msg + fmt.Sprintf(chatListAll, service)
		}

		return msg, nil
	default:
		return "", fmt.Errorf(chatIncompleteCommand, "service")
	}
}

func (rb *ReleasedBot) GroupHandler(b *BotData) (string, error) {
	if b.CommandArguments == "help" {
		return readHelpFile("group"), nil
	}

	args := strings.Fields(b.CommandArguments)
	if len(args) == 0 {
		return "", fmt.Errorf(chatIncompleteCommand, "group")
	}

	cmd := args[0]
	switch cmd {
	case "all":
		allGroup, err := rb.dbAdapter.GetAllGroup()
		if err != nil {
			return "", err
		}

		if len(allGroup) == 0 {
			return "Belum ada group yang terdaftar", nil
		}

		msg := "Daftar semua group:\n"
		for _, group := range allGroup {
			msg = msg + fmt.Sprintf(chatListAll, group)
		}

		return msg, nil
	case "add":
		if len(args) < 2 {
			return "", fmt.Errorf(chatIncompleteCommand, "group")
		}

		group := args[1]
		err := rb.dbAdapter.AddGroup(group)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf(chatGroupAdded, group), nil
	case "delete":
		if len(args) < 2 {
			return "", fmt.Errorf(chatIncompleteCommand, "group")
		}

		group := args[1]
		err := rb.dbAdapter.DeleteGroup(group)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf(chatGroupAdded, group), nil
	default:
		return "", fmt.Errorf(chatIncompleteCommand, "group")
	}
}

func readHelpFile(cmd string) string {
	fileName := "./docs/" + cmd + ".md"
	helpDoc, err := ioutil.ReadFile(fileName)
	if err != nil {
		return fmt.Sprintf(chatHelpNotFound, cmd)
	}

	return string(helpDoc)
}

func contains(arr []string, s string) bool {
	for _, v := range arr {
		if v == s {
			return true
		}
	}

	return false
}
