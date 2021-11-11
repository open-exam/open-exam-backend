package main

import (
	"context"
	"crypto/tls"
	"errors"
	"html/template"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	pb "github.com/open-exam/open-exam-backend/email-service/grpc-email-service"
	"github.com/open-exam/open-exam-backend/shared"
	userPb "github.com/open-exam/open-exam-backend/user-db-service/grpc-user-db-service"
	mail "github.com/xhit/go-simple-mail/v2"
)

type Server struct {
	pb.UnimplementedEmailServiceServer
}

func NewServer() (*Server, error) {
	return &Server{}, nil
}

type String string
func (s String) Format(data map[string]string) (out string, err error) {
    t := template.Must(template.New("").Parse(string(s)))
    builder := &strings.Builder{}
    if err = t.Execute(builder, data); err != nil {
        return "", err
    }
    out = builder.String()
    return out, nil
}

func (s *Server) SendEmail(ctx context.Context, req *pb.EmailRequest) (*pb.SendEmailResponse, error) {
	serviceResponse := make([]string, 0)

	conn, err := shared.GetGrpcConn(userService)
	if err != nil {
		return nil, err
	}

	client := userPb.NewUserServiceClient(conn)
	res, err := client.BatchGetEmails(context.Background(), &userPb.BatchGetEmailsRequest { Ids: req.Users })
	if err != nil {
		return nil, err
	}

	remove := func(slice []string, s int) []string {
		return append(slice[:s], slice[s+1:]...)
	}

	removeVar := func(slice []*pb.Variable, s int) []*pb.Variable {
		return append(slice[:s], slice[s+1:]...)
	}

	count := 0
	for i, e := range res.Emails {
		if len(e) == 0 {
			req.Users = remove(req.Users, i - count)
			req.Vars = removeVar(req.Vars, i - count)
			serviceResponse = append(serviceResponse, e)
			count++
		}
	}

	resp, err := http.Get(fsService + "/email-templates?id=" + strconv.FormatUint(req.TemplateId, 10));
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, shared.Errors.ServiceConnection
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	sb := string(body)

	if len(sb) == 0 {
		return nil, errors.New("Empty template")
	}

	server := mail.NewSMTPClient()
	server.Host = smtpHost
	server.Port = smtpPort
	server.Username = emailUser
	server.Password = emailPassword
	server.Encryption = mail.EncryptionSTARTTLS
	server.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	server.SendTimeout = 10 * time.Second
	server.ConnectTimeout = 10 * time.Second
	server.KeepAlive = true

	smtpClient, err := server.Connect()
	if err != nil {
		return nil, err
	}

	defer func() {
		start := time.Now()
		count = 0

		for i := range req.Users {

			body := sb
			body, _ = String(body).Format(req.Vars[i].Values)
	
			subject := req.Subject
			subject, _ = String(subject).Format(req.Vars[i].Values)
	
			email := mail.NewMSG()
			email.SetFrom(res.Emails[i]).SetSubject(subject).SetBody(mail.TextHTML, body)
	
			if email.Error != nil {
				continue
			}
	
			email.Send(smtpClient)
			count++

			elapsed := time.Since(start).Milliseconds()
			if count >= rateLimit && elapsed <= 1000 {
				time.Sleep(time.Duration(1000 - elapsed) * time.Millisecond)
				count = 0
				start = time.Now()
			}
		}
	}()

	return &pb.SendEmailResponse {
		Failed: serviceResponse,
	}, nil
}