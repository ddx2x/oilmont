package common

import (
	"context"
	"fmt"
	"time"

	"github.com/ddx2x/oilmont/pkg/datasource"
	"github.com/ddx2x/oilmont/pkg/log"
)

type DBNameType = string
type TableNameType = string

var DBLIst = []string{DefaultDatabase}

const (
	HttpRequestUserHeaderKey          = `x-wrapper-username`
	HttpRequestUserHeaderTENANT       = `Tenant`
	HttpRequestUserHeaderBELONGTENANT = `belong-tenant`
	AuthorizationHeader               = "Authorization"
	EXPIRETIME                        = int64(time.Hour * 6)

	// filter key
	FilterName      = "metadata.name"
	FilterNamespace = "metadata.namespace"
	FilterWorkspace = "metadata.workspace"

	// database
	DefaultDatabase DBNameType = "base"
	CustomDatabase  DBNameType = "custom"

	// kubernetes
	DefaultKubernetes = "default"
	DefaultWorkspace  = "default"

	// TABLERESOURCE
	TABLERESOURCE = "tableresource"
	// GVRRESOURCE
	GVRRESOURCE = "gvrresource"

	CLOUDEVENT = "cloudevent"

	// Vendor
	AWS    = "aws"
	ALIYUN = "aliyun"

	// cloud provider status
	INIT    string = "init"
	UPDATE  string = "update"
	RUNNING string = "running"
	DELETE  string = "delete"
	SYNC    string = "sync"
	FAIL    string = "fail"

	// IAM 基础资源
	ACCOUNT        TableNameType = "account"
	BUSINESSGROUP  TableNameType = "businessgroup"
	ROLE           TableNameType = "role"
	NAMESPACE      TableNameType = "namespace"
	PERMISSION     TableNameType = "permission"
	AUTHORITYGROUP TableNameType = "authoritygroup"
	OPERATION      TableNameType = "operation"
	RESOURCE       TableNameType = "resource"
	USER           TableNameType = "user"
	TENANT         TableNameType = "tenant"

	// IAM 连接表
	ACCOUNTPERMISSION TableNameType = "accountpermission" // 账户和权限中间表
	// 以下仅用作固定名称，不需要注册
	ACCOUNTROLE          string = "accountrole"
	ACCOUNTBUSINESSGROUP string = "accountbusinessgroup"
	BUSINESSGROUPROLE    string = "businessgrouprole"

	// Compute

	INSTANCE       TableNameType = "instance"
	IMAGE          TableNameType = "image"
	VIRTUALMACHINE TableNameType = "virtualmachine"
	VPC            TableNameType = "vpc"
	VSWITCH        TableNameType = "vswitch"
	STORAGE        TableNameType = "storage"

	// Networking
	NETWORKINTERFACE TableNameType = "networkinterface"

	// system 配置
	CLUSTER       TableNameType = "cluster"
	AVAILABLEZONE TableNameType = "availablezone"
	REGION        TableNameType = "region"
	WORKSPACE     TableNameType = "workspace"
	LICENSE       TableNameType = "license"
	PROVIDER      TableNameType = "provider"
	Menu          TableNameType = "menu"
	INSTANCETYPE  TableNameType = "instancetype"
	SECURITYGROUP TableNameType = "securitygroup"
	THEME         TableNameType = "theme"
	RELATION      TableNameType = "relation"

	// customresource 配置
	CUSTOMRESOURCE TableNameType = "customresource"
)

// 表名需要预先注册到单独的表里
var resourceTableMap = map[string]TableNameType{
	"account":           ACCOUNT,
	"businessgroup":     BUSINESSGROUP,
	"role":              ROLE,
	"namespace":         NAMESPACE,
	"permission":        PERMISSION,
	"authoritygroup":    AUTHORITYGROUP,
	"operation":         OPERATION,
	"resource":          RESOURCE,
	"accountpermission": ACCOUNTPERMISSION,
	"instance":          INSTANCE,
	"image":             IMAGE,
	"virtualmachine":    VIRTUALMACHINE,
	"vpc":               VPC,
	"cluster":           CLUSTER,
	"workspace":         WORKSPACE,
	"availablezone":     AVAILABLEZONE,
	"region":            REGION,
	"license":           LICENSE,
	"provider":          PROVIDER,
	"menu":              Menu,
	"vswitch":           VSWITCH,
	"instancetype":      INSTANCETYPE,
	"customresource":    CUSTOMRESOURCE,
	"securitygroup":     SECURITYGROUP,
	"theme":             THEME,
	"cloudevent":        CLOUDEVENT,
	"user":              USER,
	"tenant":            TENANT,
	"relation":          RELATION,
	"storage":           STORAGE,
	"networkinterface":  NETWORKINTERFACE,
}

func GetResourceTable(ctx context.Context, stage datasource.IStorage, tenant, resource string) (DBNameType, TableNameType) {
	result := make(map[string]interface{})
	if err := stage.GetById(DefaultDatabase, TABLERESOURCE, resource, result); err != nil {
		log.G(ctx).Warnf("get table not exist %s error: %s", resource, err)
	}

	table, ok := result["_id"].(string)
	if !ok {
		log.G(ctx).Warnf("get table not exist %s", resource)
	}

	db := tenant

	value, ok := result["data"].(string)
	if ok && value != "" {
		db = result["data"].(string)
	}

	if db == "" {
		db = DefaultDatabase
	}

	return db, table
}

func InsertDynCR(stage datasource.IStorage, key string, data string) error {
	return stage.InsertUnique(DefaultDatabase, TABLERESOURCE, key, data)
}

func InitResourceConfigure(stage datasource.IStorage) error {
	for key, _ := range resourceTableMap {
		if err := stage.InsertUnique(DefaultDatabase, TABLERESOURCE, key, ""); err != nil {
			return err
		}
	}
	return nil
}

var (
	// MicroServiceName
	MicroServiceName = func(name string) string { return fmt.Sprintf("/%s", name) }
	// CloudArea Identify the geographic location where the cloud application is running, default 255 is illegal value
	CloudArea = 0xff
)
