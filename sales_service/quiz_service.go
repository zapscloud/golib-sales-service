package sales_service

import (
	"fmt"
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

type QuizService interface {
	// List - List All records
	List(filter string, sort string, skip int64, limit int64) (utils.Map, error)
	// Get - Find By Code
	Get(quizId string) (utils.Map, error)
	// Find - Find the item
	Find(filter string) (utils.Map, error)
	// Create - Create Service
	Create(indata utils.Map) (utils.Map, error)
	// Update - Update Service
	Update(quizId string, indata utils.Map) (utils.Map, error)
	// Delete - Delete Service
	Delete(quizId string, delete_permanent bool) error

	EndService()
}

type quizBaseService struct {
	db_utils.DatabaseService
	dbRegion    db_utils.DatabaseService
	daoQuiz     sales_repository.QuizDao
	daoBusiness platform_repository.BusinessDao
	child       QuizService
	businessId  string
}

// NewQuizService - Construct Quiz
func NewQuizService(props utils.Map) (QuizService, error) {
	funcode := sales_common.GetServiceModuleCode() + "M" + "01"

	log.Printf("QuizService::Start ")
	// Verify whether the business id data passed
	businessId, err := utils.GetMemberDataStr(props, sales_common.FLD_BUSINESS_ID)
	if err != nil {
		return nil, err
	}

	p := quizBaseService{}
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
	p.businessId = businessId
	p.initializeService()

	_, err = p.daoBusiness.Get(businessId)
	if err != nil {
		err := &utils.AppError{
			ErrorCode:   funcode + "01",
			ErrorMsg:    "Invalid BusinessId",
			ErrorDetail: "Given BusinessId is not exist"}
		return p.errorReturn(err)
	}

	p.child = &p

	return &p, err
}

// quizBaseService - Close all the services
func (p *quizBaseService) EndService() {
	log.Printf("EndService ")
	p.CloseDatabaseService()
	p.dbRegion.CloseDatabaseService()
}

func (p *quizBaseService) initializeService() {
	log.Printf("QuizService:: GetBusinessDao ")
	p.daoBusiness = platform_repository.NewBusinessDao(p.GetClient())
	p.daoQuiz = sales_repository.NewQuizDao(p.GetClient(), p.businessId)
}

// List - List All records
func (p *quizBaseService) List(filter string, sort string, skip int64, limit int64) (utils.Map, error) {

	log.Println("quizBaseService::FindAll - Begin")

	listdata, err := p.daoQuiz.List(filter, sort, skip, limit)
	if err != nil {
		return nil, err
	}

	log.Println("quizBaseService::FindAll - End ")
	return listdata, nil
}

// Get - Find By Code
func (p *quizBaseService) Get(quizId string) (utils.Map, error) {
	log.Printf("quizBaseService::Get::  Begin %v", quizId)

	data, err := p.daoQuiz.Get(quizId)

	log.Println("quizBaseService::Get:: End ", err)
	return data, err
}

func (p *quizBaseService) Find(filter string) (utils.Map, error) {
	fmt.Println("QuizService::FindByCode::  Begin ", filter)

	data, err := p.daoQuiz.Find(filter)
	log.Println("QuizService::FindByCode:: End ", err)
	return data, err
}

// Create - Create Service
func (p *quizBaseService) Create(indata utils.Map) (utils.Map, error) {

	log.Println("QuizService::Create - Begin")
	var quizId string

	dataval, dataok := indata[sales_common.FLD_QUIZ_ID]
	if dataok {
		quizId = strings.ToLower(dataval.(string))
	} else {
		quizId = utils.GenerateUniqueId("quiz")
		log.Println("Unique Quiz ID", quizId)
	}

	// Assign BusinessId
	indata[sales_common.FLD_BUSINESS_ID] = p.businessId
	indata[sales_common.FLD_QUIZ_ID] = quizId

	data, err := p.daoQuiz.Create(indata)
	if err != nil {
		return utils.Map{}, err
	}

	log.Println("QuizService::Create - End ")
	return data, nil
}

// Update - Update Service
func (p *quizBaseService) Update(quizId string, indata utils.Map) (utils.Map, error) {

	log.Println("QuizService::Update - Begin")

	data, err := p.daoQuiz.Update(quizId, indata)

	log.Println("QuizService::Update - End ")
	return data, err
}

// Delete - Delete Service
func (p *quizBaseService) Delete(quizId string, delete_permanent bool) error {

	log.Println("QuizService::Delete - Begin", quizId)

	if delete_permanent {
		result, err := p.daoQuiz.Delete(quizId)
		if err != nil {
			return err
		}
		log.Printf("Delete %v", result)
	} else {
		indata := utils.Map{db_common.FLD_IS_DELETED: true}
		data, err := p.Update(quizId, indata)
		if err != nil {
			return err
		}
		log.Println("Update for Delete Flag", data)
	}

	log.Printf("QuizService::Delete - End")
	return nil
}

func (p *quizBaseService) errorReturn(err error) (QuizService, error) {
	// Close the Database Connection
	p.EndService()
	return nil, err
}
