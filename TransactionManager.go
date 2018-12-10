package GoMybatis

import (
	"errors"
	"log"
)

type Transaction_Status int

const (
	Transaction_Status_NO       Transaction_Status = iota //非事务
	Transaction_Status_Prepare                            //准备事务
	Transaction_Status_Commit                             //提交事务
	Transaction_Status_Rollback                           //回滚事务
)

type ActionType int

const (
	ActionType_Exec  ActionType = iota //执行
	ActionType_Query                   //查询
)

type TransactionReqDTO struct {
	Status        Transaction_Status
	TransactionId string //事务id(不可空)
	OwnerId       string //所有者
	Sql           string //sql内容(可空)
	ActionType    ActionType
}

type TransactionRspDTO struct {
	TransactionId string //事务id(不可空)
	Error         string
	Success       int
	Query         []map[string][]byte
	Exec          Result
}

type TransactionManager interface {
	GetTransaction(def *TransactionDefinition, transactionId string, OwnerId string) (*TransactionStatus, error)
	Commit(transactionId string) error
	Rollback(transactionId string) error
}

type DefaultTransationManager struct {
	TransactionManager
	SessionFactory     *SessionFactory
	TransactionFactory *TransactionFactory
}

func (this DefaultTransationManager) New(SessionFactory *SessionFactory, TransactionFactory *TransactionFactory) DefaultTransationManager {
	this.SessionFactory = SessionFactory
	this.TransactionFactory = TransactionFactory
	return this
}

func (this DefaultTransationManager) GetTransaction(def *TransactionDefinition, transactionId string, OwnerId string) (*TransactionStatus, error) {
	if transactionId == "" {
		return nil, errors.New("[TransactionManager] transactionId =" + transactionId + " transations is nil!")
	}
	if def == nil {
		var d = TransactionDefinition{}.Default()
		def = &d
	}
	var transationStatus, err = this.TransactionFactory.GetTransactionStatus(transactionId)
	if err != nil {
		return nil, err
	}
	if def.PropagationBehavior == PROPAGATION_REQUIRED {
		//todo doBegin
		if transationStatus.IsNewTransaction {
			//新事务，则调用begin
			transationStatus.OwnerId = OwnerId
			var err = transationStatus.Begin()
			if err == nil {
				if def.Timeout != 0 {
					//transation out of time,default not set out of time
					//事务超时,时间大于0则启动超时机制
					transationStatus.DelayFlush(def.Timeout)
				}
				return transationStatus, err
			}
		}
	}
	return transationStatus, nil
}

func (this DefaultTransationManager) Commit(transactionId string) error {
	var transactions, err = this.TransactionFactory.GetTransactionStatus(transactionId)
	if err != nil {
		log.Println(err)
		return err
	}
	return transactions.Commit()
}

func (this DefaultTransationManager) Rollback(transactionId string) error {
	var transactions, err = this.TransactionFactory.GetTransactionStatus(transactionId)
	if err != nil {
		log.Println(err)
		return err
	}
	return transactions.Rollback()
}

//执行事务
func (this DefaultTransationManager) DoTransaction(dto TransactionReqDTO) TransactionRspDTO {
	var transcationStatus *TransactionStatus
	var err error

	transcationStatus, err = this.GetTransaction(nil, dto.TransactionId, dto.OwnerId)
	if transcationStatus == nil || transcationStatus.Transaction == nil || transcationStatus.Transaction.Session == nil {
		return TransactionRspDTO{
			TransactionId: dto.TransactionId,
			Error:         "Transaction does not exist,id=" + dto.TransactionId,
		}
	}
	if err != nil {
		return TransactionRspDTO{
			TransactionId: dto.TransactionId,
			Error:         err.Error(),
		}
	}
	if err != nil {
		return TransactionRspDTO{
			TransactionId: dto.TransactionId,
			Error:         err.Error(),
		}
	}
	log.Println("[TransactionManager] do transactionId=", dto.TransactionId,",sessionId=",transcationStatus.Transaction.Session.Id())

	if dto.Status == Transaction_Status_NO {
		defer transcationStatus.Flush() //关闭
		return this.DoAction(dto, transcationStatus)
	} else if dto.Status == Transaction_Status_Prepare {
		return this.DoAction(dto, transcationStatus)
	} else if dto.Status == Transaction_Status_Commit {
		if transcationStatus.OwnerId == dto.OwnerId { //PROPAGATION_REQUIRED 情况下 子事务 不可提交
			defer transcationStatus.Flush() //关闭
			err = transcationStatus.Commit()
			if err != nil {
				return TransactionRspDTO{
					TransactionId: dto.TransactionId,
					Error:         err.Error(),
				}
			}
			var transaction, err = this.TransactionFactory.GetTransactionStatus(dto.TransactionId)
			if err != nil {
				log.Println(err)
			} else {
				transaction.Flush()
			}
		}
	} else if dto.Status == Transaction_Status_Rollback {
		defer transcationStatus.Flush() //关闭，//PROPAGATION_REQUIRED 情况下 子事务 可关闭
		err = transcationStatus.Rollback()
		if err != nil {
			return TransactionRspDTO{
				TransactionId: dto.TransactionId,
				Error:         err.Error(),
			}
		}
	} else {
		err = errors.New("[TransactionManager] arg have no action!")
	}
	var errString = ""
	if err != nil {
		errString = err.Error()
	}
	return TransactionRspDTO{
		TransactionId: dto.TransactionId,
		Error:         errString,
	}
}

//执行数据库操作
func (this DefaultTransationManager) DoAction(dto TransactionReqDTO, transcationStatus *TransactionStatus) TransactionRspDTO {
	if transcationStatus.IsCompleted {
		var TransactionRspDTO = TransactionRspDTO{
			TransactionId: dto.TransactionId,
			Error:         "[TransactionManager] transaction fail!it is completed!",
		}
		return TransactionRspDTO
	}
	if dto.Sql == "" {
		var TransactionRspDTO = TransactionRspDTO{
			TransactionId: dto.TransactionId,
		}
		return TransactionRspDTO
	}
	if dto.ActionType == ActionType_Exec {
		log.Println("[TransactionManager] TransactionId:",dto.TransactionId,",Exec:", dto.Sql)
		var res, e = transcationStatus.Transaction.Session.Exec(dto.Sql)
		var err string
		if e != nil {
			err = e.Error()
			return TransactionRspDTO{
				TransactionId: dto.TransactionId,
				Error:         err,
			}
		} else {
			return TransactionRspDTO{
				TransactionId: dto.TransactionId,
				Exec:          *res,
				Error:         err,
			}
		}
	} else {
		log.Println("[TransactionManager] Query ", dto.Sql)
		var res, e = transcationStatus.Transaction.Session.Query(dto.Sql)
		var err string
		if e != nil {
			err = e.Error()
		}
		var TransactionRspDTO = TransactionRspDTO{
			TransactionId: dto.TransactionId,
			Query:         res,
			Error:         err,
		}
		return TransactionRspDTO
	}
}
