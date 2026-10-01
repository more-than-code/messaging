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

	"encoding/base64"

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
	SmsProvider     string `envconfig:"SMS_PROVIDER"`
	EmailProvider   string `envconfig:"EMAIL_PROVIDER"`
	EmailDomains    string `envconfig:"EMAIL_DOMAINS"`
	EmailBypassCode string `envconfig:"EMAIL_BYPASS_CODE"`
	PhoneBypassCode string `envconfig:"PHONE_BYPASS_CODE"`
	IsDev           bool   `envconfig:"IS_DEV"`
	ServerPort      string `envconfig:"SERVER_PORT"`
	ProductName     string `envconfig:"PRODUCT_NAME"`
	// PostmarkAPIKey 供未傳入郵件配置的舊版客戶端使用，不得寫入日誌。
	PostmarkAPIKey string `envconfig:"POSTMARK_API_KEY"`
	// PostmarkMailSender 為舊版部署已在 Postmark 驗證的寄件地址。
	PostmarkMailSender string `envconfig:"POSTMARK_MAIL_SENDER"`
}

type Server struct {
	smsVendor sms.SmsVendor
	repo      *repository.Repository
	cfg       *ServerConfig
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

	repo, err := repository.NewRepository()
	if err != nil {
		return err
	}

	log.Printf("messaging server starting gRPC listener on %s", cfg.ServerPort)
	pb.RegisterMessagingServer(grpcServer, &Server{smsVendor: smsVendor, repo: repo, cfg: &cfg})
	err = grpcServer.Serve(lis)

	if err != nil {
		log.Printf("messaging server stopped with error: %v", err)
		return err
	}

	log.Printf("messaging server stopped on %s", cfg.ServerPort)

	return nil
}

// GenerateVerificationCode 依 req 的收件地址與模板發送驗證碼，使用 ctx 存取快取；回傳發送狀態或失敗原因。
func (s *Server) GenerateVerificationCode(ctx context.Context, req *pb.GenerateVerificationCodeRequest) (*pb.GenerateVerificationCodeResponse, error) {
	channel := "sms"
	if util.IsEmail(req.PhoneOrEmail) {
		channel = "email"
	}
	// 僅記錄通道與執行階段，避免將收件地址、驗證碼或請求憑證寫入發送流程日誌。
	log.Printf("[GenerateVerificationCode] start channel=%s", channel)
	res := &pb.GenerateVerificationCodeResponse{Status: pb.VerificationCodeGenerationStatus_VERIFICATION_CODE_GENERATION_STATUS_DONE, Msg: string(constant.MsgDone)}
	var err error

	found, err := s.repo.GetVerificationInfo(ctx, req.PhoneOrEmail)

	if err != nil {
		log.Printf("[GenerateVerificationCode] failed stage=read_verification_info error=%v", err)
		return nil, err
	}

	// 先檢查重發間隔，避免重複呼叫外部供應商與產生額外費用。
	if found != nil {
		if time.Since(found.LastAttempt).Minutes() <= 1 {
			log.Printf("[GenerateVerificationCode] rate_limited channel=%s", channel)

			res.Status = pb.VerificationCodeGenerationStatus_VERIFICATION_CODE_GENERATION_STATUS_SENDING_TOO_FREQUENTLY
			res.Msg = string(constant.MsgSendingTooFrequently)

			return res, nil
		}
	}

	code := strconv.Itoa(rand.Intn(9000) + 1000)

	message, err := templateToMessage(req.MessageTemplate, code)

	if err != nil {
		log.Printf("[GenerateVerificationCode] failed stage=render_template error=%v", err)
		return nil, err
	}

	// 郵件供應商依新舊配置解析；電話維持既有簡訊供應商流程。
	if channel == "email" {
		mailVendor, mailErr := s.resolveEmailVendor(req.EmailConfig)
		if mailErr != nil {
			log.Printf("[GenerateVerificationCode] failed stage=email_config error=%v", mailErr)
			return nil, mailErr
		}
		err = mailVendor.SendCode(req.PhoneOrEmail, req.Subject, message)
	} else {
		err = s.smsVendor.SendCodeNProduct(req.PhoneOrEmail, code, s.cfg.ProductName)
	}

	if err != nil {
		log.Printf("[GenerateVerificationCode] failed stage=send channel=%s error=%v", channel, err)
		return nil, err
	}

	// 供應商確認發送後才保存驗證碼；寫入失敗須獨立記錄，因為訊息可能已送達。
	ph := repository.VerificationInfo{Code: code, Attempt: 0, LastAttempt: time.Now()}

	err = s.repo.SetVerificationInfo(ctx, req.PhoneOrEmail, &ph)

	if err != nil {
		log.Printf("[GenerateVerificationCode] failed stage=save_verification_info channel=%s error=%v", channel, err)
		return nil, err
	}

	log.Printf("[GenerateVerificationCode] done channel=%s", channel)

	return res, nil
}

func (s *Server) ValidateVerificationCode(ctx context.Context, req *pb.ValidateVerificationCodeRequest) (*pb.ValidateVerificationCodeResponse, error) {
	log.Printf("[ValidateVerificationCode] START - PhoneOrEmail: %s, Code: %s", req.PhoneOrEmail, req.VerificationCode)

	var msg = constant.MsgValid
	var status = pb.VerificationCodeValidationStatus_VERIFICATION_CODE_VALIDATION_STATUS_VALID

	// if s.cfg.IsDev {
	// 	log.Printf("[ValidateVerificationCode] IsDev mode enabled, returning VALID directly")
	// 	return &pb.ValidateVerificationCodeResponse{Status: status, Msg: string(msg)}, nil
	// }

	// Check email bypass code
	emailDomains := strings.Split(s.cfg.EmailDomains, ",")
	userDomain := util.DomainFromAddress(req.PhoneOrEmail)
	if util.Contains(emailDomains, userDomain) && req.VerificationCode == s.cfg.EmailBypassCode {
		log.Printf("[ValidateVerificationCode] Email bypass code matched for domain: %s", userDomain)
		return &pb.ValidateVerificationCodeResponse{Status: status, Msg: string(msg)}, nil
	}

	// Check phone bypass code
	if !util.IsEmail(req.PhoneOrEmail) && req.VerificationCode == s.cfg.PhoneBypassCode {
		log.Printf("[ValidateVerificationCode] Phone bypass code matched")
		return &pb.ValidateVerificationCodeResponse{Status: status, Msg: string(msg)}, nil
	}

	found, err := s.repo.GetVerificationInfo(ctx, req.PhoneOrEmail)
	if err != nil {
		log.Printf("[ValidateVerificationCode] ERROR getting verification info: %v", err)
		return nil, err
	}

	log.Printf("[ValidateVerificationCode] Found in Redis: %v", found != nil)
	if found != nil {
		log.Printf("[ValidateVerificationCode] Stored Code: %s, Request Code: %s, Attempt: %d", found.Code, req.VerificationCode, found.Attempt)

		if found.Code == req.VerificationCode {
			log.Printf("[ValidateVerificationCode] Code MATCHED - returning VALID")
			s.repo.DeleteVerificationInfo(ctx, strings.ToLower(req.PhoneOrEmail))
		} else {
			log.Printf("[ValidateVerificationCode] Code MISMATCH")
			if found.Attempt >= 3 {
				log.Printf("[ValidateVerificationCode] Maximum attempts reached")
				msg = constant.MsgMaximumAttempts
				status = pb.VerificationCodeValidationStatus_VERIFICATION_CODE_VALIDATION_STATUS_MAXIMUM_ATTEMPTS
				s.repo.DeleteVerificationInfo(ctx, strings.ToLower(req.PhoneOrEmail))
			} else {
				log.Printf("[ValidateVerificationCode] Invalid code, incrementing attempt to %d", found.Attempt+1)
				msg = constant.MsgInvalid
				status = pb.VerificationCodeValidationStatus_VERIFICATION_CODE_VALIDATION_STATUS_INVALID
				found.Attempt++
				s.repo.SetVerificationInfo(ctx, req.PhoneOrEmail, found)
			}
		}
	} else {
		log.Printf("[ValidateVerificationCode] No verification info found - code EXPIRED")
		msg = constant.MsgExpired
		status = pb.VerificationCodeValidationStatus_VERIFICATION_CODE_VALIDATION_STATUS_EXPIRED
	}

	log.Printf("[ValidateVerificationCode] FINAL RESULT - Status: %v, Msg: %s", status, msg)
	return &pb.ValidateVerificationCodeResponse{Status: status, Msg: string(msg)}, nil
}

// SendEmailWithAttachment 處理 ctx 所屬 RPC，依 req 的郵件配置與附件發送郵件，回傳成功狀態或失敗原因。
func (s *Server) SendEmailWithAttachment(ctx context.Context, req *pb.SendEmailWithAttachmentRequest) (*pb.SendEmailWithAttachmentResponse, error) {
	log.Printf("[SendEmailWithAttachment] start")
	attachments := []email.Attachment{}
	if req.Attachment != nil {
		// RPC 附件為原始位元組，供應商介面要求 Base64，須先轉換再發送。
		encoded := base64.StdEncoding.EncodeToString(req.Attachment.Content)
		attachments = append(attachments, email.Attachment{Name: req.Attachment.Name, Content: encoded, ContentType: "application/octet-stream"})
	}

	mailVendor, err := s.resolveEmailVendor(req.EmailConfig)
	if err != nil {
		log.Printf("[SendEmailWithAttachment] failed stage=email_config error=%v", err)
		return nil, err
	}

	// 與驗證碼共用配置解析，讓舊版附件郵件呼叫也能使用部署環境配置。
	err = mailVendor.SendEmailWithAttachment(req.To, req.Bcc, req.Subject, req.Message, attachments)

	if err != nil {
		log.Printf("[SendEmailWithAttachment] failed stage=send error=%v", err)
		return nil, err
	}

	log.Printf("[SendEmailWithAttachment] done")
	return &pb.SendEmailWithAttachmentResponse{Success: true}, nil
}

// resolveEmailVendor 優先使用 cfg；僅 cfg 缺省時採用舊版 Postmark 環境配置，回傳供應商或配置錯誤。
func (s *Server) resolveEmailVendor(cfg *pb.EmailConfig) (email.EmailVendor, error) {
	source := "request"
	if cfg == nil {
		source = "environment"
		if s.cfg == nil {
			return nil, fmt.Errorf("legacy postmark server config is required")
		}
		// 舊版客戶端未定義 EmailConfig，沿用原先固定使用 Postmark 的部署契約。
		cfg = &pb.EmailConfig{
			Provider:    string(email.ProviderPostmark),
			ApiKey:      s.cfg.PostmarkAPIKey,
			EmailSender: s.cfg.PostmarkMailSender,
		}
	}
	log.Printf("[resolveEmailVendor] source=%s provider=%q", source, cfg.Provider)

	// 明確傳入的空配置或錯誤憑證必須報錯，禁止回退至其他寄件帳號。
	emailCfg, err := translateEmailConfig(cfg)
	if err != nil {
		return nil, err
	}

	return email.NewVendor(emailCfg)
}

func translateEmailConfig(cfg *pb.EmailConfig) (email.Config, error) {
	if cfg == nil {
		return email.Config{}, fmt.Errorf("email config is required")
	}

	provider := email.ProviderType(strings.ToUpper(cfg.Provider))
	if provider == "" {
		return email.Config{}, fmt.Errorf("email provider is required")
	}

	switch provider {
	case email.ProviderPostmark, email.ProviderMailchimp:
		result := email.Config{Provider: provider, APIKey: cfg.ApiKey, EmailSender: cfg.EmailSender}
		return result, nil
	default:
		return email.Config{}, fmt.Errorf("unsupported email provider: %s", cfg.Provider)
	}
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
