package messaging

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/keighl/postmark"
	"github.com/more-than-code/messaging/constant"
	"github.com/more-than-code/messaging/email-vendor"
	"github.com/more-than-code/messaging/pb"
	"github.com/more-than-code/messaging/repository"
	"github.com/more-than-code/messaging/sms-vendor"

	"github.com/more-than-code/messaging/util"

	"github.com/kelseyhightower/envconfig"
	"google.golang.org/grpc"
)

type ServerConfig struct {
	SmsProvider   string `envconfig:"SMS_PROVIDER"`
	EmailProvider string `envconfig:"EMAIL_PROVIDER"`
	EmailDomains  string `envconfig:"EMAIL_DOMAINS"`
	BypassCode    string `envconfig:"BYPASS_CODE"`
	IsDev         bool   `envconfig:"IS_DEV"`
	ServerPort    string `envconfig:"SERVER_PORT"`
	ProductName   string `envconfig:"PRODUCT_NAME"`
}

type Server struct {
	smsVendor  sms.SmsVendor
	mailVendor email.EmailVendor
	repo       *repository.Repository
	cfg        *ServerConfig
	pb.UnimplementedMessagingServer
}

func NewServer() error {
	var cfg ServerConfig
	err := envconfig.Process("", &cfg)

	if err != nil {
		log.Fatal(err)
	}

	lis, err := net.Listen("tcp", cfg.ServerPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	opts = append(opts, grpc.MaxRecvMsgSize(1024*1024*10))

	grpcServer := grpc.NewServer(opts...)
	var smsVendor sms.SmsVendor

	switch cfg.SmsProvider {
	case "VOLC":
		smsVendor, err = sms.NewVolcVendor()
	case "BYTEPLUS":
		smsVendor, err = sms.NewBytePlusVendor()
	default:
		log.Fatal("Invalid SMS provider")
	}

	if err != nil {
		return err
	}

	mailVendor, err := email.NewVendor()
	if err != nil {
		return err
	}

	repo, err := repository.NewRepository()
	if err != nil {
		return err
	}

	pb.RegisterMessagingServer(grpcServer, &Server{smsVendor: smsVendor, mailVendor: mailVendor, repo: repo, cfg: &cfg})
	err = grpcServer.Serve(lis)

	if err != nil {
		return err
	}

	return nil
}

func (s *Server) GenerateVerificationCode(ctx context.Context, req *pb.GenerateVerificationCodeRequest) (*pb.GenerateVerificationCodeResponse, error) {
	res := &pb.GenerateVerificationCodeResponse{Status: pb.VerificationCodeGenerationStatus_VERIFICATION_CODE_GENERATION_STATUS_DONE, Msg: string(constant.MsgDone)}
	var err error

	found, err := s.repo.GetVerificationInfo(ctx, req.PhoneOrEmail)

	if err != nil {
		return nil, err
	}

	if found != nil {
		if time.Since(found.LastAttempt).Minutes() <= 1 {
			str := fmt.Sprintf("Code sent for %s within 1 minute", req.PhoneOrEmail)
			fmt.Println(str)

			res.Status = pb.VerificationCodeGenerationStatus_VERIFICATION_CODE_GENERATION_STATUS_SENDING_TOO_FREQUENTLY
			res.Msg = string(constant.MsgSendingTooFrequently)

			return res, nil
		}
	}

	code := strconv.Itoa(rand.Intn(9000) + 1000)

	message, err := templateToMessage(req.MessageTemplate, code)

	if err != nil {
		return nil, err
	}

	if util.IsEmail(req.PhoneOrEmail) {
		err = s.mailVendor.SendCode(req.PhoneOrEmail, req.Subject, message)
	} else {
		err = s.smsVendor.SendCodeNProduct(req.PhoneOrEmail, code, s.cfg.ProductName)
	}

	if err != nil {
		return nil, err
	}

	ph := repository.VerificationInfo{Code: code, Attempt: 0, LastAttempt: time.Now()}

	err = s.repo.SetVerificationInfo(ctx, req.PhoneOrEmail, &ph)

	if err != nil {
		return nil, err
	}

	log.Println("Sent to " + req.PhoneOrEmail + " with code " + code)

	return res, nil
}

func (s *Server) ValidateVerificationCode(ctx context.Context, req *pb.ValidateVerificationCodeRequest) (*pb.ValidateVerificationCodeResponse, error) {
	var msg = constant.MsgValid
	var status = pb.VerificationCodeValidationStatus_VERIFICATION_CODE_VALIDATION_STATUS_VALID

	if s.cfg.IsDev {
		return &pb.ValidateVerificationCodeResponse{Status: status, Msg: string(msg)}, nil
	}

	if (util.Contains(strings.Split(s.cfg.EmailDomains, ","), util.DomainFromAddress(req.PhoneOrEmail)) || !util.IsEmail(req.PhoneOrEmail)) && req.VerificationCode == s.cfg.BypassCode {
		return &pb.ValidateVerificationCodeResponse{Status: status, Msg: string(msg)}, nil
	}

	found, err := s.repo.GetVerificationInfo(ctx, req.PhoneOrEmail)

	if err != nil {
		return nil, err
	}

	if found != nil {
		if found.Code == req.VerificationCode {
			s.repo.DeleteVerificationInfo(ctx, req.PhoneOrEmail)
		} else {
			if found.Attempt >= 3 {
				msg = constant.MsgMaximumAttempts
				status = pb.VerificationCodeValidationStatus_VERIFICATION_CODE_VALIDATION_STATUS_MAXIMUM_ATTEMPTS
				s.repo.DeleteVerificationInfo(ctx, req.PhoneOrEmail)
			} else if found.Code != req.VerificationCode {
				msg = constant.MsgInvalid
				status = pb.VerificationCodeValidationStatus_VERIFICATION_CODE_VALIDATION_STATUS_INVALID
				found.Attempt++
				s.repo.SetVerificationInfo(ctx, req.PhoneOrEmail, found)
			}
		}

	} else {
		msg = constant.MsgExpired
		status = pb.VerificationCodeValidationStatus_VERIFICATION_CODE_VALIDATION_STATUS_EXPIRED
	}

	return &pb.ValidateVerificationCodeResponse{Status: status, Msg: string(msg)}, nil
}

func (s *Server) SendEmailWithAttachment(ctx context.Context, req *pb.SendEmailWithAttachmentRequest) (*pb.SendEmailWithAttachmentResponse, error) {
	attachments := []postmark.Attachment{}
	if req.Attachment != nil {
		attachments = append(attachments, postmark.Attachment{Name: req.Attachment.Name, Content: string(req.Attachment.Content), ContentType: "application/octet-stream"})
	}

	err := s.mailVendor.SendEmailWithAttachment(req.To, req.Bcc, req.Subject, req.Message, attachments)

	if err != nil {
		return nil, err
	}

	return &pb.SendEmailWithAttachmentResponse{}, nil
}

func templateToMessage(msgTemplate string, code string) (string, error) {
	tmpl, err := template.New("message").Parse(msgTemplate)

	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, struct{ Code string }{code})

	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
