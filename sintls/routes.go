package sintls

import (
	"github.com/gin-gonic/gin"
	"github.com/go-pg/pg"
	"golang.org/x/crypto/bcrypt"
	"github.com/go-acme/lego/challenge"
	"log"
	"net/http"
	"github.com/zehome/sintls/dns"
)

/* Lego httpreq RAW request */
type LegoMessage struct {
	Domain  string `json:"domain"`
	Token   string `json:"token"`
	KeyAuth string `json:"keyAuth"`
	TargetA string `json:"dnstarget_a"`
	TargetAAAA string `json:"dnstarget_aaaa"`
	TargetCNAME string `json:"dnstarget_cname"`
}

func updateDNSRecords(db *pg.Tx, req LegoMessage, user *Authorization, dnsupdater *dns.DNSUpdater) error {
	err = user.CreateOrUpdateHost(
		tx, req.Domain, req.TargetA, req.TargetAAAA, req.TargetCNAME)
	if err != nil {
		tx.Rollback()
		log.Printf("sintls: update host failed: %s", err)
		return
	}
	// Update DNS records
	if len(req.TargetA) > 0 {
		err = dnsupdater.SetRecord(req.Domain, "A", req.TargetA)
		if err != nil {
			log.Printf("sintls: setrecord A failed: %s", err)
			return
		}
	}
	if len(req.TargetAAAA) > 0 {
		err = dnsupdater.SetRecord(req.Domain, "AAAA", req.TargetAAAA)
		if err != nil {
			log.Printf("sintls: setrecord AAAA failed: %s", err)
			return
		}
	}
	if len(req.TargetA) == 0 && len(req.TargetAAAA) == 0 && len(req.TargetCNAME) > 0 {
		err = dnsupdater.SetRecord(req.Domain, "CNAME", req.TargetCNAME)
		return
	}
	return
}


func LegoPresent(c *gin.Context) {
	var req LegoMessage
	if err := c.ShouldBindJSON(&req); err != nil {
		c.String(http.StatusBadRequest, "bad request:", err.Error())
		return
	}
	user := c.MustGet("user").(*Authorization)
	db := c.MustGet("database").(*pg.DB)
	if user == nil {
		c.String(http.StatusUnauthorized, "user auth is required")
		return
	}
	if ! user.CanUseHost(db, req.Domain) {
		c.String(http.StatusForbidden, "no permissions to use this domain")
		return
	}
	// Lego DNS Provider
	provider := c.MustGet("dnsprovider").(challenge.Provider)
	// Custom DNS Updater
	dnsupdater := c.MustGet("dnsupdater").(dns.DNSUpdater)

	tx, err := db.Begin()
	if err != nil {
		log.Printf("sintls: db.Begin() failed", err)
		c.String(http.StatusInternalServerError, "begin failed")
		return
	}

	err = updateDNSRecords(tx, req, user, &dnsupdater)
	if err != nil {
		tx.Rollback()
		c.String(http.StatusInternalServerError, "update dnsrecords failed")
		return
	}

	// Lego challenge
	err = provider.Present(req.Domain, req.Token, req.KeyAuth)
	if err != nil {
		tx.Rollback()
		log.Printf("sintls: %s", err)
		c.String(http.StatusInternalServerError, "lego present failed")
		return
	}

	if err := tx.Commit(); err != nil {
		log.Printf("sintls: tx.Commit() failed: ", err)
		c.String(http.StatusInternalServerError, "commit failed")
	} else {
		c.String(http.StatusOK, "success")
	}
}

func LegoCleanup(c *gin.Context) {
	var req LegoMessage
	if err := c.ShouldBindJSON(&req); err != nil {
		c.String(http.StatusBadRequest, "bad request:", err.Error())
		return
	}
	user := c.MustGet("user").(*Authorization)
	db := c.MustGet("database").(*pg.DB)
	if user == nil {
		c.String(http.StatusUnauthorized, "user auth is required")
		return
	}
	if ! user.CanUseHost(db, req.Domain) {
		c.String(http.StatusForbidden, "no permissions to use this domain")
		return
	}
	provider := c.MustGet("dnsprovider").(challenge.Provider)
	err := provider.CleanUp(req.Domain, req.Token, req.KeyAuth)
	if err != nil {
		log.Printf("sintls: %s", err)
		c.String(http.StatusInternalServerError, "lego Cleanup failed")
		return
	}
	c.String(http.StatusOK, "success")
}

type CreateAuthMessage struct {
	Name   string `json:"name" binding:"required"`
	Secret string `json:"secret" binding:"required"`
}

func CreateAuth(c *gin.Context) {
	var authmessage CreateAuthMessage
	user := c.MustGet("user").(*Authorization)
	if user.Admin.Bool != true {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	db := c.MustGet("database").(*pg.DB)
	if err := c.ShouldBindJSON(&authmessage); err != nil {
		c.String(http.StatusBadRequest, "bad request: %s", err.Error())
		return
	}
	var count int = 0
	count, err := db.Model((*Authorization)(nil)).
		Where("name = ?", authmessage.Name).
		Count()
	if count != 0 || err != nil {
		c.String(http.StatusForbidden, "Authorization %s already exists", authmessage.Name)
		return
	}
	hashpw, err := bcrypt.GenerateFromPassword(
		[]byte(authmessage.Secret), bcrypt.DefaultCost)
	if err != nil {
		c.String(http.StatusBadRequest, "unhashable password: %s", err)
		return
	}
	dbauth := Authorization{
		Name:   authmessage.Name,
		Secret: string(hashpw),
	}
	if err := db.Insert(&dbauth); err != nil {
		c.String(http.StatusInternalServerError, "Unknown error: %s", err)
		return
	}
	c.String(http.StatusOK, "success")
}

type DeleteAuthMessage struct {
	Name string `json:"name" binding:"required"`
}

func DeleteAuth(c *gin.Context) {
	var req DeleteAuthMessage
	//var auth Authorization
	user := c.MustGet("user").(*Authorization)
	if user.Admin.Bool != true {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	db := c.MustGet("database").(*pg.DB)
	if err := c.ShouldBindJSON(&req); err != nil {
		c.String(http.StatusBadRequest, "bad request: %s", err)
		return
	}
	tx, err := db.Begin()
	if err != nil {
		log.Printf("db.Begin() failed: %s", err)
		return
	}
	// Rollback tx on error.
	defer tx.Rollback()

	// tx := db.Begin()
	// if tx.Error != nil {
	// 	log.Fatal("db.Begin() failed", tx.Error)
	// 	c.String(http.StatusInternalServerError, "begin failed")
	// 	return
	// }
	// if err := db.Where("name = ?", req.Name).First(&auth).Error; err != nil {
	// 	c.String(http.StatusNotFound, "user %s not found", req.Name)
	// 	return
	// }
	// if err := db.Model(&auth).Related(&subdomains).Error; err == nil {
	// 	for _, subdomain := range subdomains {
	// 		if err := db.Model(&subdomain).Related(&hosts).Error; err == nil {
	// 			for _, host := range hosts {
	// 				db.Delete(&host)
	// 			}
	// 		}
	// 	}
	// }

	// if err := tx.Commit().Error; err != nil {
	// 	log.Fatal("tx.Commit() failed: ", err)
	// 	c.String(http.StatusInternalServerError, "commit failed")
	// } else {
	c.String(http.StatusOK, "success")
	//}
}
