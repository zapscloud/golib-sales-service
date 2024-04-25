package sales_service

import (
	"log"
	"strings"

	"github.com/zapscloud/golib-dbutils/db_common"
	"github.com/zapscloud/golib-dbutils/db_utils"
	"github.com/zapscloud/golib-platform-repository/platform_repository"
	"github.com/zapscloud/golib-platform-service/platform_service"
	"github.com/zapscloud/golib-sales-repository/sales_common"
	"github.com/zapscloud/golib-sales-repository/sales_repository"
	"github.com/zapscloud/golib-utils/utils"
)

// OTPService - States Service structure
type OTPService interface {
	List(filter string, sort string, skip int64, limit int64) (utils.Map, error)
	Get(otp_id string) (utils.Map, error)
	Find(filter string) (utils.Map, error)
	Create(indata utils.Map) (utils.Map, error)
	Update(otp_id string, indata utils.Map) (utils.Map, error)
	Delete(otp_id string, delete_permanent bool) error
	Verify(key, otp string) (utils.Map, error)

	BeginTransaction()
	CommitTransaction()
	RollbackTransaction()

	EndService()
}

// OTPBaseService - States Service structure
type OTPBaseService struct {
	db_utils.DatabaseService
	dbRegion            db_utils.DatabaseService
	daoOTP              sales_repository.OTPDao
	daoPlatformBusiness platform_repository.BusinessDao
	child               OTPService
	businessID          string
}

func init() {
	log.SetFlags(log.Lshortfile | log.LstdFlags | log.Lmicroseconds)
}

func NewOTPService(props utils.Map) (OTPService, error) {
	funcode := sales_common.GetServiceModuleCode() + "M" + "01"

	log.Printf("OTPService::Start ")

	// Verify whether the business id data passed
	businessId, err := utils.GetMemberDataStr(props, sales_common.FLD_BUSINESS_ID)
	if err != nil {
		return nil, err
	}

	p := OTPBaseService{}
	// Open Database Service
	err = p.OpenDatabaseService(props)
	if err != nil {
		return nil, err
	}

	// Open RegionDB Service
	p.dbRegion, err = platform_service.OpenRegionDatabaseService(props)
	if err != nil {
		p.CloseDatabaseService()
		return nil, err
	}

	// Assign the BusinessId
	p.businessID = businessId

	// Instantiate other services
	p.daoOTP = sales_repository.NewOTPDao(p.dbRegion.GetClient(), p.businessID)
	p.daoPlatformBusiness = platform_repository.NewBusinessDao(p.GetClient())

	_, err = p.daoPlatformBusiness.Get(p.businessID)
	if err != nil {
		err := &utils.AppError{
			ErrorCode:   funcode + "01",
			ErrorMsg:    "Invalid business id",
			ErrorDetail: "Given business id is not exist"}
		return p.errorReturn(err)
	}

	p.child = &p

	return &p, nil
}

func (p *OTPBaseService) EndService() {
	p.CloseDatabaseService()
	p.dbRegion.CloseDatabaseService()
}

// List - List All records
func (p *OTPBaseService) List(filter string, sort string, skip int64, limit int64) (utils.Map, error) {

	log.Println("OTPService::FindAll - Begin")

	daoOTP := p.daoOTP
	response, err := daoOTP.List(filter, sort, skip, limit)
	if err != nil {
		return nil, err
	}

	log.Println("OTPService::FindAll - End ")
	return response, nil
}

// FindByCode - Find By Code
func (p *OTPBaseService) Get(otp_id string) (utils.Map, error) {
	log.Printf("OTPService::FindByCode::  Begin %v", otp_id)

	data, err := p.daoOTP.Get(otp_id)
	log.Println("OTPService::FindByCode:: End ", err)
	return data, err
}

func (p *OTPBaseService) Find(filter string) (utils.Map, error) {
	log.Println("OTPService::FindByCode::  Begin ", filter)

	data, err := p.daoOTP.Find(filter)
	log.Println("OTPService::FindByCode:: End ", data, err)
	return data, err
}

func (p *OTPBaseService) Create(indata utils.Map) (utils.Map, error) {

	log.Println("UserService::Create - Begin")

	var otp_id string

	dataval, dataok := indata[sales_common.FLD_OTP_ID]
	if dataok {
		otp_id = strings.ToLower(dataval.(string))
	} else {
		otp_id = utils.GenerateUniqueId("otp")
		log.Println("Unique State ID", otp_id)
	}
	indata[sales_common.FLD_OTP_ID] = otp_id
	indata[sales_common.FLD_BUSINESS_ID] = p.businessID
	log.Println("Provided State ID:", otp_id)

	_, err := p.daoOTP.Get(otp_id)
	if err == nil {
		err := &utils.AppError{ErrorCode: "S30102", ErrorMsg: "Existing State ID !", ErrorDetail: "Given State ID already exist"}
		return indata, err
	}

	insertResult, err := p.daoOTP.Create(indata)
	if err != nil {
		return indata, err
	}
	log.Println("UserService::Create - End ", insertResult)
	return indata, err
}

// Update - Update Service
func (p *OTPBaseService) Update(otp_id string, indata utils.Map) (utils.Map, error) {

	log.Println("OTPService::Update - Begin")

	data, err := p.daoOTP.Get(otp_id)
	if err != nil {
		return data, err
	}

	// Delete key fields
	delete(indata, sales_common.FLD_OTP_ID)
	delete(indata, sales_common.FLD_BUSINESS_ID)

	data, err = p.daoOTP.Update(otp_id, indata)
	log.Println("OTPService::Update - End ")
	return data, err
}

// Delete - Delete Service
func (p *OTPBaseService) Delete(otp_id string, delete_permanent bool) error {

	log.Println("OTPService::Delete - Begin", otp_id)

	daoOTP := p.daoOTP
	_, err := daoOTP.Get(otp_id)
	if err != nil {
		return err
	}

	if delete_permanent {
		result, err := daoOTP.Delete(otp_id)
		if err != nil {
			return err
		}
		log.Printf("Delete %v", result)
	} else {
		indata := utils.Map{db_common.FLD_IS_DELETED: true}
		data, err := daoOTP.Update(otp_id, indata)
		if err != nil {
			return err
		}
		log.Println("Update for Delete Flag", data)
	}

	log.Printf("OTPService::Delete - End")
	return nil
}

// Delete - Delete Service
func (p *OTPBaseService) Verify(key, otp string) (utils.Map, error) {
	log.Println("OTPService::Verify - Begin", key, otp)
	data, err := p.daoOTP.Verify(key, otp)
	log.Println("OTPService::Verify - End", data, err)
	return data, err
}

func (p *OTPBaseService) errorReturn(err error) (OTPService, error) {
	// Close the Database Connection
	p.EndService()
	return nil, err
}
