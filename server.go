package email

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/more-than-code/messaging/pb"
	"github.com/more-than-code/messaging/repository"
	"github.com/more-than-code/messaging/sender"

	"github.com/more-than-code/messaging/constant"
	"github.com/more-than-code/messaging/util"

	"github.com/kelseyhightower/envconfig"
	"google.golang.org/grpc"
)

type ServerConfig struct {
	EmailDomains string `envconfig:"EMAIL_DOMAINS"`
	BypassCode   string `envconfig:"BYPASS_CODE"`
	AppName      string `envconfig:"APP_NAME"`
}

type Server struct {
	smsVendor  *sender.SmsVendor
	mailVendor *sender.MailVendor
	repo       *repository.Repository
	cfg        *ServerConfig
	pb.UnimplementedEmailServer
}

func NewServer(port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption

	grpcServer := grpc.NewServer(opts...)
	smsVendor, err := sender.NewSmsVendor()
	if err != nil {
		return err
	}

	mailVendor, err := sender.NewMailVendor()
	if err != nil {
		return err
	}

	repo, err := repository.NewRepository()
	if err != nil {
		return err
	}

	var cfg ServerConfig
	err = envconfig.Process("", &cfg)

	if err != nil {
		log.Fatal(err)
	}

	pb.RegisterEmailServer(grpcServer, &Server{smsVendor: smsVendor, mailVendor: mailVendor, repo: repo, cfg: &cfg})
	err = grpcServer.Serve(lis)

	if err != nil {
		return err
	}

	return nil
}

func (s *Server) GenerateVerificationCode(ctx context.Context, req *pb.GenerateVerificationCodeRequest) (*pb.GenerateVerificationCodeResponse, error) {
	res := &pb.GenerateVerificationCodeResponse{Code: 0, Msg: "Done"}
	var err error

	found, err := s.repo.GetVerificationInfo(ctx, req.PhoneOrEmail)

	if err != nil {
		return nil, err
	}

	if found != nil {
		if time.Since(found.LastAttempt).Minutes() <= 1 {
			str := fmt.Sprintf("Code sent for %s within 1 minute", req.PhoneOrEmail)
			fmt.Println(str)

			res.Code = constant.ErrCodeTooFrequentlySendingVerificationCode
			res.Msg = constant.MsgTooFrequentlySendingVerificationCode

			return res, nil
		}
	}

	code := strconv.Itoa(rand.Intn(9000) + 1000)

	if util.IsEmail(req.PhoneOrEmail) {
		err = s.mailVendor.SendCodeFromPostmark(req.PhoneOrEmail, s.cfg.AppName+constant.SmsSubject, fmt.Sprintf("%s%s:<strong>%s</strong>", s.cfg.AppName, purposeToMessage(req.Purpose), code))
	} else {

		err = s.smsVendor.SendCodeGlobe(req.PhoneOrEmail, s.cfg.AppName+purposeToMessage(req.Purpose)+": "+code)

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

func (s *Server) VerifyCode(ctx context.Context, req *pb.VerifyCodeRequest) (*pb.VerifyCodeResponse, error) {
	var errMsg = "Verified"
	var errCode = 0

	if util.Contains(strings.Split(s.cfg.EmailDomains, ","), util.DomainFromAddress(req.PhoneOrEmail)) && req.VerificationCode == s.cfg.BypassCode {
		return &pb.VerifyCodeResponse{Code: int32(errCode), Msg: errMsg}, nil
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
				errMsg = constant.MsgMaximumAttemptsOnVerificationCode
				errCode = constant.ErrCodeMaximumAttemptsOnVerificationCode
				s.repo.DeleteVerificationInfo(ctx, req.PhoneOrEmail)
			} else if found.Code != req.VerificationCode {
				errMsg = constant.MsgWrongVerificationCode
				errCode = constant.ErrCodeWrongVerificationCode

				found.Attempt++
				s.repo.SetVerificationInfo(ctx, req.PhoneOrEmail, found)
			}
		}

	} else {
		errMsg = constant.MsgExpiredVerificationCode
		errCode = constant.ErrCodeExpiredVerificationCode
	}

	return &pb.VerifyCodeResponse{Code: int32(errCode), Msg: errMsg}, nil
}

func purposeToMessage(purpose int32) string {
	switch purpose {
	case constant.PurposeRegistration:
		return constant.SmsRegistration
	case constant.PurposeLogin:
		return constant.SmsLogin
	case constant.PurposeResettingPassword:
		return constant.SmsResettingPassword
	case constant.PurposeBindingPhoneOrEmail:
		return constant.SmsBindingPhoneOrEmail
	default:
		return ""
	}
}
