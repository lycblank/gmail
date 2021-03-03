package main

import (
	"flag"
	"fmt"
	"github.com/lycblank/gmail"
	"github.com/tealeg/xlsx"
	ggmail "google.golang.org/api/gmail/v1"
	"os"
	"regexp"
	"github.com/lycblank/goprogressbar"
	"runtime"
	"sync"
	"sync/atomic"
)

var subjectQuery string
var savePath string
var maxResult int64

func main() {
	ParseArgs()
	config, err := gmail.GetOauth2Config("credentials.json")
	if err != nil {
		panic(err)
	}
	client := gmail.GetClient(config, "token.json")
	svr, err := ggmail.New(client)
	if err != nil {
		panic(err)
	}
	ll := svr.Users.Messages.List("me")
	r, err := ll.MaxResults(maxResult).Q(fmt.Sprintf("subject:%s", subjectQuery)).Do()
	if err != nil {
		panic(err)
	}

	wb := xlsx.NewFile()
	sh, err := wb.AddSheet(subjectQuery)
	if err != nil {
		panic(err)
	}
	row := sh.AddRow()
	row.AddCell().Value = "from"
	row.AddCell().Value = "subject"
	row.AddCell().Value = "email_id"

	bar := progressbar.NewProgressBar(int64(len(r.Messages)))
	cpuNum := runtime.NumCPU()
	group := sync.WaitGroup{}
	group.Add(cpuNum)

	ch := make(chan string, cpuNum)
	for i:=0;i<cpuNum;i++{
		go func() {
			defer group.Done()
			for emailId := range ch {
				ProcessMessage(svr, sh, emailId, bar)
			}
		}()
	}

	for _, message := range r.Messages {
		ch <- message.Id
	}
	close(ch)
	group.Wait()

	bar.Finish()
	wb.Save(savePath)
}

func init() {
	flag.StringVar(&subjectQuery, "s", "", "搜索邮件的subject")
	flag.StringVar(&savePath, "o", "result.xlsx", "保存结果路径")
	flag.Int64Var(&maxResult, "m", 1000, "搜索的最大结果")
}

func ParseArgs() {
	flag.Parse()
	if subjectQuery == "" {
		flag.Usage()
		os.Exit(1)
	}
	if savePath == "" {
		flag.Usage()
		os.Exit(1)
	}
	if maxResult <= 0 {
		flag.Usage()
		os.Exit(1)
	}
}

func GetCSVMessage(message *ggmail.Message) CSVMessage {
	re:=regexp.MustCompile(`[a-zA-Z0-9]+@.+\.[a-zA-Z]+`)
	msg := CSVMessage{}
	for i, cnt := 0,len(message.Payload.Headers); i < cnt;i++ {
		if message.Payload.Headers[i].Name == "From" {
			match:=re.FindString(message.Payload.Headers[i].Value)
			msg.From = match
		}
		if message.Payload.Headers[i].Name == "Subject" {
			msg.Subject = message.Payload.Headers[i].Value
		}
		if msg.From != "" && msg.Subject != "" {
			break
		}
	}
	return msg
}

func GetGmailMessage(svr *ggmail.Service, emailId string) *ggmail.Message {
	msg := svr.Users.Messages.Get("me", emailId)
	cc, err := msg.Format("METADATA").Do()
	if err != nil {
		return nil
	}
	return cc
}

var sheetMutex sync.Mutex
func AddXLSXRecord(sheet *xlsx.Sheet, msg CSVMessage) {
	sheetMutex.Lock()
	row := sheet.AddRow()
	sheetMutex.Unlock()
	row.AddCell().Value = msg.From
	row.AddCell().Value = msg.Subject
	row.AddCell().Value = msg.EmailId
}

func ProcessMessage(svr *ggmail.Service, sheet *xlsx.Sheet, emailId string, bar *progressbar.ProgressBar) {
	defer AddProcessBar(bar, 1)

	gmailMsg := GetGmailMessage(svr, emailId)
	if gmailMsg == nil {
		return
	}
	csvMsg := GetCSVMessage(gmailMsg)
	csvMsg.EmailId = emailId

	AddXLSXRecord(sheet, csvMsg)


}

var processValue int64
func AddProcessBar(bar *progressbar.ProgressBar, step int64) {
	atomic.AddInt64(&processValue, step)
	bar.Play(atomic.LoadInt64(&processValue))
}

type CSVMessage struct {
	From string
	Subject string
	EmailId string
}
