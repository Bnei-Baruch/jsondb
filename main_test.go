// main_test.go

package main_test

import (
	"os"
	"testing"

	"."
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
)

var a main.App

func TestMain(m *testing.M) {
	a = main.App{}
	a.Initialize(
		os.Getenv("TEST_DB_USERNAME"),
		os.Getenv("TEST_DB_PASSWORD"),
		os.Getenv("TEST_DB_NAME"))

	ensureTableExists()

	code := m.Run()

	clearTable()

	os.Exit(code)
}

func ensureTableExists() {
	if _, err := a.DB.Exec(tableCreationQuery); err != nil {
		log.Fatal(err)
	}
}

func clearTable() {
	a.DB.Exec("DELETE FROM products")
	a.DB.Exec("ALTER SEQUENCE products_id_seq RESTART WITH 1")
}

const tableCreationQuery = `CREATE TABLE IF NOT EXISTS products
(
id SERIAL,
name TINYTEXT NOT NULL,
price NUMERIC(10,2) NOT NULL DEFAULT 0.00,
CONSTRAINT products_pkey PRIMARY KEY (id)
)`

const createCaptureTable = `CREATE TABLE IF NOT EXISTS capture
(
id BIGSERIAL,
capture_id TEXT NOT NULL,
capture_src TEXT,
date VARCHAR(10),
start_name TEXT,
stop_name TEXT,
sha1 VARCHAR(40),
line jsonb,
original jsonb,
proxy jsonb,
wfstatus jsonb,
CONSTRAINT capture_pkey PRIMARY KEY (capture_id)
)`

const createIngestTable = `CREATE TABLE IF NOT EXISTS ingest
(
id BIGSERIAL,
capture_id TEXT NOT NULL,
capture_src TEXT,
date VARCHAR(10),
start_name TEXT,
stop_name TEXT,
sha1 VARCHAR(40),
line jsonb,
original jsonb,
proxy jsonb,
wfstatus jsonb,
CONSTRAINT ingest_pkey PRIMARY KEY (capture_id)
)`

const createStateTable = `CREATE TABLE IF NOT EXISTS state
(
id SERIAL,
state_id TEXT NOT NULL,
data jsonb NOT NULL,
CONSTRAINT state_pkey PRIMARY KEY (state_id)
)`

const createTrimmerTable = `CREATE TABLE IF NOT EXISTS trimmer
(
id BIGSERIAL,
trim_id TEXT NOT NULL,
date VARCHAR(10),
file_name TEXT,
inpoints REAL[],
outpoints REAL[],
parent jsonb,
line jsonb,
original jsonb,
proxy jsonb,
wfstatus jsonb,
CONSTRAINT trimmer_pkey PRIMARY KEY (trim_id)
)`

const createDgimaTable = `CREATE TABLE IF NOT EXISTS dgima
(
id BIGSERIAL,
dgima_id TEXT NOT NULL,
date VARCHAR(10),
file_name TEXT,
inpoints REAL[],
outpoints REAL[],
parent jsonb,
line jsonb,
original jsonb,
proxy jsonb,
wfstatus jsonb,
CONSTRAINT dgima_pkey PRIMARY KEY (dgima_id)
)`

const createArichaTable = `CREATE TABLE IF NOT EXISTS aricha
(
id BIGSERIAL,
aricha_id TEXT NOT NULL,
date VARCHAR(10),
file_name TEXT,
parent jsonb,
line jsonb,
original jsonb,
proxy jsonb,
wfstatus jsonb,
CONSTRAINT aricha_pkey PRIMARY KEY (aricha_id)
)`

const createJobsTable = `CREATE TABLE IF NOT EXISTS jobs
(
id BIGSERIAL,
job_id TEXT NOT NULL,
date VARCHAR(10),
file_name TEXT,
job_name TEXT,
job_type TEXT NOT NULL,
parent jsonb,
line jsonb,
original jsonb,
proxy jsonb,
product jsonb,
wfstatus jsonb,
CONSTRAINT job_pkey PRIMARY KEY (job_id)
)`

const createArchiveTable = `CREATE TABLE IF NOT EXISTS archive
(
id BIGSERIAL,
archive_id TEXT NOT NULL,
created_datetime timestamp NOT NULL DEFAULT now(),
date VARCHAR(10) NOT NULL,
file_name VARCHAR NOT NULL,
language VARCHAR(10) NOT NULL,
extension VARCHAR(10) NOT NULL,
source VARCHAR(20) DEFAULT NULL,
send_id TEXT DEFAULT NULL,
size BIGINT NOT NULL,
sha1 VARCHAR(40) NOT NULL,
CONSTRAINT archive_pkey PRIMARY KEY (archive_id)
)`

const createUConvertTable = `CREATE TABLE IF NOT EXISTS convert
(
id BIGSERIAL,
convert_id TEXT NOT NULL,
name TEXT,
date VARCHAR(10),
progress TEXT,
state TEXT,
timestamp TEXT,
langcheck jsonb,
CONSTRAINT convert_pkey PRIMARY KEY (convert_id)
)`

const createCarbonTable = `CREATE TABLE IF NOT EXISTS carbon
(
id BIGSERIAL,
carbon_id TEXT NOT NULL,
send_id TEXT NOT NULL,
date VARCHAR(10) NOT NULL,
file_name VARCHAR NOT NULL,
language VARCHAR(10) NOT NULL,
extension VARCHAR(10) NOT NULL,
size BIGINT NOT NULL,
duration REAL NOT NULL,
sha1 VARCHAR(40) NOT NULL,
CONSTRAINT carbon_pkey PRIMARY KEY (carbon_id)
)`

const createUploadTable = `CREATE TABLE IF NOT EXISTS kmedia
(
id BIGSERIAL,
kmedia_id TEXT NOT NULL,
date VARCHAR(10) NOT NULL,
file_name VARCHAR NOT NULL,
language VARCHAR(10) NOT NULL,
extension VARCHAR(10) NOT NULL,
source VARCHAR(20) DEFAULT NULL,
send_id TEXT DEFAULT NULL,
size BIGINT NOT NULL,
sha1 VARCHAR(40) NOT NULL,
pattern TEXT,
CONSTRAINT kmedia_pkey PRIMARY KEY (kmedia_id)
)`

const createUInsertTable = `CREATE TABLE IF NOT EXISTS insert
(
id BIGSERIAL,
insert_id TEXT NOT NULL,
insert_name TEXT NOT NULL,
date VARCHAR(10) NOT NULL,
file_name VARCHAR NOT NULL,
extension VARCHAR(10) NOT NULL,
size BIGINT NOT NULL,
sha1 VARCHAR(40) NOT NULL,
language VARCHAR(10) NOT NULL,
insert_type VARCHAR(1) NOT NULL,
send_id TEXT,
upload_type TEXT NOT NULL,
line jsonb NOT NULL,
CONSTRAINT insert_pkey PRIMARY KEY (insert_id)
)`

const createUFilesTable = `CREATE TABLE IF NOT EXISTS files
(
id BIGSERIAL,
file_id TEXT NOT NULL,
date VARCHAR(10) NOT NULL,
file_name VARCHAR NOT NULL,
extension VARCHAR(10) NOT NULL,
size BIGINT NOT NULL,
sha1 VARCHAR(40) NOT NULL,
file_type TEXT NOT NULL,
send_id TEXT,
line jsonb NOT NULL,
CONSTRAINT files_pkey PRIMARY KEY (sha1)
)`

const createUStateTable = `CREATE TABLE IF NOT EXISTS state
(
id BIGSERIAL,
state_id TEXT NOT NULL,
tag TEXT,
data jsonb NOT NULL,
CONSTRAINT state_pkey PRIMARY KEY (state_id)
)`

func TestEmptyTable(t *testing.T) {
	clearTable()

	req, _ := http.NewRequest("GET", "/products", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	if body := response.Body.String(); body != "[]" {
		t.Errorf("Expected an empty array. Got %s", body)
	}
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func TestGetNonExistentProduct(t *testing.T) {
	clearTable()

	req, _ := http.NewRequest("GET", "/product/11", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNotFound, response.Code)

	var m map[string]string
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["error"] != "Not Found" {
		t.Errorf("Expected the 'error' key of the response to be set to 'Not Found'. Got '%s'", m["error"])
	}
}

func TestCreateProduct(t *testing.T) {
	clearTable()

	payload := []byte(`{"name":"test product","price":11.22}`)

	req, _ := http.NewRequest("POST", "/product", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["name"] != "test product" {
		t.Errorf("Expected product name to be 'test product'. Got '%v'", m["name"])
	}

	if m["price"] != 11.22 {
		t.Errorf("Expected product price to be '11.22'. Got '%v'", m["price"])
	}

	// the id is compared to 1.0 because JSON unmarshaling converts numbers to
	// floats, when the target is a map[string]interface{}
	if m["id"] != 1.0 {
		t.Errorf("Expected product ID to be '1'. Got '%v'", m["id"])
	}
}

func TestGetProduct(t *testing.T) {
	clearTable()
	addProducts(1)

	req, _ := http.NewRequest("GET", "/product/1", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
}

func addProducts(count int) {
	if count < 1 {
		count = 1
	}

	for i := 0; i < count; i++ {
		a.DB.Exec("INSERT INTO products(name, price) VALUES($1, $2)", "Product "+strconv.Itoa(i), (i+1.0)*10)
	}
}

func TestUpdateProduct(t *testing.T) {
	clearTable()
	addProducts(1)

	req, _ := http.NewRequest("GET", "/product/1", nil)
	response := executeRequest(req)
	var originalProduct map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &originalProduct)

	payload := []byte(`{"name":"test product - updated name","price":11.22}`)

	req, _ = http.NewRequest("PUT", "/product/1", bytes.NewBuffer(payload))
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)

	if m["id"] != originalProduct["id"] {
		t.Errorf("Expected the id to remain the same (%v). Got %v", originalProduct["id"], m["id"])
	}

	if m["name"] == originalProduct["name"] {
		t.Errorf("Expected the name to change from '%v' to '%v'. Got '%v'", originalProduct["name"], m["name"], m["name"])
	}

	if m["price"] == originalProduct["price"] {
		t.Errorf("Expected the price to change from '%v' to '%v'. Got '%v'", originalProduct["price"], m["price"], m["price"])
	}
}

func TestDeleteProduct(t *testing.T) {
	clearTable()
	addProducts(1)

	req, _ := http.NewRequest("GET", "/product/1", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("DELETE", "/product/1", nil)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/product/1", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}
