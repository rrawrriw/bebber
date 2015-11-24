package bebber

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"gopkg.in/mgo.v2/bson"
)

func Test_FindDoc_OK(t *testing.T) {
	globals := MakeTestGlobals(t)
	session := globals.MongoDB.Session.Copy()
	defer session.Close()

	db := session.DB(TestDBName)
	defer db.DropDatabase()

	docID := bson.NewObjectId()
	expectDoc := Doc{
		ID:     docID,
		Name:   "karl.pdf",
		Note:   "Note",
		Labels: []Label{},
		AccountData: DocAccountData{
			DocNumbers: []string{},
		},
	}
	err := db.C(DocsColl).Insert(expectDoc)
	if err != nil {
		t.Fatal(err.Error())
	}

	doc := Doc{Name: "karl.pdf"}
	err = doc.Find(db)
	if err != nil {
		t.Fatal(err.Error())
	}

	expectDocJSON, err := json.Marshal(expectDoc)
	if err != nil {
		t.Fatal(err.Error())
	}

	docJSON, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err.Error())
	}

	if string(expectDocJSON) != string(docJSON) {
		t.Fatal("Exepect", string(expectDocJSON), "was", string(docJSON))
	}

}

func Test_FindDoc_Fail(t *testing.T) {
	globals := MakeTestGlobals(t)
	session := globals.MongoDB.Session.Copy()
	defer session.Close()

	db := session.DB(TestDBName)
	defer db.DropDatabase()

	doc := Doc{Name: "karl.pdf"}
	err = doc.Find(db)
	if strings.Contains(err.Error(), "Cannot find") == false {
		t.Fatal("Expect to fail with Cannot find document was", err)
	}
}

func Test_ChangeDoc_OK(t *testing.T) {
	globals := MakeTestGlobals(t)
	session := globals.MongoDB.Session.Copy()
	defer session.Close()

	db := session.DB(TestDBName)
	defer db.DropDatabase()

	dateOfScan := time.Now()
	dateOfReceipt := dateOfScan
	doc := Doc{
		Name: "changeme",
		Infos: DocInfos{
			DateOfScan:    dateOfScan,
			DateOfReceipt: dateOfReceipt,
		},
		AccountData: DocAccountData{
			DocNumbers: []string{"2"},
			AccNumber:  1223,
		},
	}

	docsColl := db.C(DocsColl)
	err := docsColl.Insert(doc)
	if err != nil {
		t.Fatal(err.Error())
	}

	d := time.Now()
	newDateOfReceipt := dateOfReceipt.AddDate(0, 0, 2)
	accountData := DocAccountData{
		DocNumbers: []string{"1"},
		DocPeriod: DocPeriod{
			From: d,
			To:   d,
		},
	}
	labels := []Label{"label1"}
	changeDoc := Doc{
		Infos: DocInfos{
			DateOfReceipt: newDateOfReceipt,
		},
		Name:        "nicer",
		Barcode:     "barcode",
		AccountData: accountData,
		Labels:      labels,
		Note:        "note",
	}
	err = doc.Change(changeDoc, db)
	if err != nil {
		t.Fatal(err.Error())
	}

	failDoc := Doc{Name: "changeme"}
	err = failDoc.Find(db)
	if err == nil {
		t.Fatal("Expect Cannot find document error was nil")
	}
	if strings.Contains(err.Error(), "Cannot find document") == false {
		t.Fatal("Expect Cannot find document error was", err.Error())
	}

	docUpdated := Doc{Name: "nicer"}
	err = docUpdated.Find(db)
	if err != nil {
		t.Fatal(err.Error())
	}

	if doc.Name != "nicer" {
		t.Fatal("Expect nicer was", doc.Name)
	}

	if doc.Barcode != "barcode" {
		t.Fatal("Expect barcode was", doc.Barcode)
	}

	if docUpdated.Barcode != "barcode" {
		t.Fatal("Expect barcode was", docUpdated.Barcode)
	}

	if docUpdated.Note != "note" {
		t.Fatal("Expect note was", docUpdated.Note)
	}

	if CompareDates(docUpdated.Infos.DateOfScan, dateOfScan) == false {
		t.Fatal("Expect", dateOfScan, "was", docUpdated.Infos.DateOfScan)
	}

	if CompareDates(docUpdated.Infos.DateOfReceipt, newDateOfReceipt) == false {
		t.Fatal("Exepct", newDateOfReceipt, "was", docUpdated.Infos.DateOfReceipt)
	}

	if len(docUpdated.AccountData.DocNumbers) != 1 {
		t.Fatal("Expect 1 was", docUpdated.AccountData.DocNumbers)
	}

	if docUpdated.AccountData.AccNumber != 1223 {
		t.Fatal("Expect 1223 was", docUpdated.AccountData.AccNumber)
	}

	if docUpdated.AccountData.DocPeriod.To.IsZero() {
		t.Fatal("Expect", d, "was", docUpdated.AccountData.DocPeriod.To)
	}
	if docUpdated.Labels[0] != "label1" {
		t.Fatal("Expect label1 was", docUpdated.Labels[0])
	}
}

func Test_ChangeDoc_DateOfScanFail(t *testing.T) {
	globals := MakeTestGlobals(t)
	session := globals.MongoDB.Session.Copy()
	defer session.Close()

	db := session.DB(TestDBName)
	defer db.DropDatabase()

	doc := Doc{Name: "changeme", Infos: DocInfos{}}

	docsColl := db.C(DocsColl)
	err := docsColl.Insert(doc)
	if err != nil {
		t.Fatal(err.Error())
	}

	docInfos := DocInfos{
		DateOfScan: time.Now(),
	}
	changeDoc := Doc{Name: "changeme", Infos: docInfos}
	err = doc.Change(changeDoc, db)
	errMsg := "Not allowed to change date of scan!"
	if strings.Contains(err.Error(), errMsg) == false {
		t.Fatal("Expect", errMsg, "was", err.Error())
	}
}

func Test_RemoveDoc_OK(t *testing.T) {
	globals := MakeTestGlobals(t)
	session := globals.MongoDB.Session.Copy()
	defer session.Close()

	db := session.DB(TestDBName)
	defer db.DropDatabase()

	doc := Doc{Name: "Pimpel", Infos: DocInfos{}}

	docsColl := db.C(DocsColl)
	err := docsColl.Insert(doc)
	if err != nil {
		t.Fatal(err.Error())
	}

	err = doc.Remove(db)

	if err != nil {
		t.Fatal(err.Error())
	}

	err = doc.Find(db)
	if strings.Contains(err.Error(), "Cannot find document") == false {
		t.Fatal(err.Error())
	}

}

func Test_AppendLabels_OK(t *testing.T) {
	globals := MakeTestGlobals(t)
	session := globals.MongoDB.Session.Copy()
	defer session.Close()

	db := session.DB(TestDBName)
	defer db.DropDatabase()

	docTmp := Doc{
		Name:   "Hoocker",
		Labels: []Label{"label1"},
	}
	docsColl := db.C(DocsColl)
	err := docsColl.Insert(docTmp)
	if err != nil {
		t.Fatal(err.Error())
	}

	labels := []Label{Label("l2"), Label("l3")}
	docTmp.AppendLabels(labels, db)

	doc := Doc{Name: "Hoocker"}
	err = doc.Find(db)
	if err != nil {
		t.Fatal(err)
	}

	if len(doc.Labels) != 3 {
		t.Fatal("Expect 3 labels was", doc.Labels)
	}
	if len(docTmp.Labels) != 3 {
		t.Fatal("Expect 3 labels was", doc.Labels)
	}
}

func Test_AppendLabel_DoupleFail(t *testing.T) {
	globals := MakeTestGlobals(t)
	session := globals.MongoDB.Session.Copy()
	defer session.Close()

	db := session.DB(TestDBName)
	defer db.DropDatabase()

	docTmp := Doc{
		Name:   "Hoocker",
		Labels: []Label{"single"},
	}
	docsColl := db.C(DocsColl)
	err := docsColl.Insert(docTmp)
	if err != nil {
		t.Fatal(err.Error())
	}

	labels := []Label{Label("l2"), Label("single")}
	err = docTmp.AppendLabels(labels, db)

	errMsg := "Label already exists!"
	if strings.Contains(err.Error(), errMsg) == false {
		t.Fatal("Expect", errMsg, "was", err)
	}

}

func Test_RemoveLabels_OK(t *testing.T) {
	globals := MakeTestGlobals(t)
	session := globals.MongoDB.Session.Copy()
	defer session.Close()

	db := session.DB(TestDBName)
	defer db.DropDatabase()

	docTmp := Doc{
		Name:   "Hoocker",
		Labels: []Label{"l1", "l2", "l3"},
	}
	docsColl := db.C(DocsColl)
	err := docsColl.Insert(docTmp)
	if err != nil {
		t.Fatal(err.Error())
	}

	labels := []Label{"l2", "l3"}
	docTmp.RemoveLabels(labels, db)

	doc := Doc{Name: "Hoocker"}
	err = doc.Find(db)
	if err != nil {
		t.Fatal(err)
	}

	if len(doc.Labels) != 1 {
		t.Fatal("Expect 1 labels was", doc.Labels)
	}
	if len(docTmp.Labels) != 1 {
		t.Fatal("Expect 1 labels was", doc.Labels)
	}

	if docTmp.Labels[0] != "l1" {
		t.Fatal("Expect l1 was", docTmp.Labels[0])
	}
}

func Test_AppendDocNumbers_OK(t *testing.T) {
	globals := MakeTestGlobals(t)
	session := globals.MongoDB.Session.Copy()
	defer session.Close()

	db := session.DB(TestDBName)
	defer db.DropDatabase()

	docTmp := Doc{
		Name: "Hoocker",
		AccountData: DocAccountData{
			DocNumbers: []string{"1"},
		},
	}
	docsColl := db.C(DocsColl)
	err := docsColl.Insert(docTmp)
	if err != nil {
		t.Fatal(err.Error())
	}

	numbers := []string{"2", "3"}
	docTmp.AppendDocNumbers(numbers, db)

	doc := Doc{Name: "Hoocker"}
	err = doc.Find(db)
	if err != nil {
		t.Fatal(err)
	}

	if len(doc.AccountData.DocNumbers) != 3 {
		t.Fatal("Expect 3 labels was", doc.AccountData.DocNumbers)
	}
	if len(docTmp.AccountData.DocNumbers) != 3 {
		t.Fatal("Expect 3 labels was", docTmp.AccountData.DocNumbers)
	}
}

func Test_RemoveDocNumbers_OK(t *testing.T) {
	globals := MakeTestGlobals(t)
	session := globals.MongoDB.Session.Copy()
	defer session.Close()

	db := session.DB(TestDBName)
	defer db.DropDatabase()

	docTmp := Doc{
		Name: "Hoocker",
		AccountData: DocAccountData{
			DocNumbers: []string{"1", "2", "3"},
		},
	}
	docsColl := db.C(DocsColl)
	err := docsColl.Insert(docTmp)
	if err != nil {
		t.Fatal(err.Error())
	}

	docNumbers := []string{"1", "2"}
	docTmp.RemoveDocNumbers(docNumbers, db)

	doc := Doc{Name: "Hoocker"}
	err = doc.Find(db)
	if err != nil {
		t.Fatal(err)
	}

	if len(doc.AccountData.DocNumbers) != 1 {
		t.Fatal("Expect 1 labels was", doc.AccountData.DocNumbers)
	}
	if len(docTmp.AccountData.DocNumbers) != 1 {
		t.Fatal("Expect 1 labels was", doc.AccountData.DocNumbers)
	}
	if docTmp.AccountData.DocNumbers[0] != "3" {
		t.Fatal("Expect 3 was", docTmp.AccountData.DocNumbers[0])
	}
}
