package holochain

import (
	_ "fmt"
	"os"
	"io/ioutil"
	"errors"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"github.com/google/uuid"
	"github.com/BurntSushi/toml"
)

const Version string = "0.0.1"

const (
	DirectoryName string = ".holochain"
	DNAFileName string = "dna.conf"
	ChanFileName string = "chain.db"
	LocalFileName string = "local.conf"
	SysFileName string = "system.conf"
	PubKeyFileName string = "pub.key"
	PrivKeyFileName string = "priv.key"
)

type Config struct {
	Port string
}

type Holochain struct {
	Id uuid.UUID
	LinkEncoding string
}

func IsInitialized(path string) bool {
	return dirExists(path+"/"+DirectoryName) && fileExists(path+"/"+DirectoryName+"/"+SysFileName)
}

func IsConfigured(path string) bool {
	return fileExists(path+"/"+DNAFileName)
}

// New creates a new holochain structure with a randomly generated ID and default values
func New() Holochain {
	u,err := uuid.NewUUID()
	if err != nil {panic(err)}
	return Holochain {Id:u,LinkEncoding:"JSON"}
}

// Load creates a holochain structure from the configuration files
func Load(path string) (h Holochain,err error) {
	if IsConfigured(path) {
		_,err = toml.DecodeFile(path+"/"+DNAFileName, &h)
	} else {
		err = mkErr("missing "+DNAFileName)
	}
	return h,err
}

// GenChain setts up a holochain by creating the initial genesis links.
// It assumes a properly set up .holochain sub-directory with a config file and
// keys for signing.  See Gen
func GenChain() (err error) {
	return errors.New("not implemented")
}


//Init initializes service defaults and a new key pair in the dirname directory
func Init(path string) error {
	if err := os.MkdirAll(path+"/"+DirectoryName,os.ModePerm); err != nil {
		return err
	}
	c := Config {}
	err := writeToml(path,SysFileName,c)
	if err != nil {return err}

	priv,err := ecdsa.GenerateKey(elliptic.P256(),rand.Reader)
	if err != nil {return err}

	mpriv,err := x509.MarshalECPrivateKey(priv)
	if err != nil {return err}

	err = writeFile(path,PrivKeyFileName,mpriv)
	if err != nil {return err}

	var pub *ecdsa.PublicKey
	pub = priv.Public().(*ecdsa.PublicKey)
	mpub,err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {return err}

	err = writeFile(path,PubKeyFileName,mpub)
	return err
}

// Gen adds template files suitable for editing to the given path
func GenDev(path string) (hP *Holochain, err error) {
	var h Holochain
	if err := os.MkdirAll(path,os.ModePerm); err != nil {
		return nil,err;
	}

	h = New()

	if err = writeToml(path,DNAFileName,h); err == nil {
		hP = &h
	}

	return
}
/*func Link(h *Holochain, data interface{}) error {

}*/

func writeToml(path string,file string,data interface{}) error {
	p := path+"/"+file
	if fileExists(p) {
		return mkErr(path+" already exists")
	}
	f, err := os.Create(p)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := toml.NewEncoder(f)
	err = enc.Encode(data);
	return err
}

func writeFile(path string,file string,data []byte) error {
	p := path+"/"+file
	if fileExists(p) {
		return mkErr(path+" already exists")
	}
	f, err := os.Create(p)
	if err != nil {return err}

	defer f.Close()
	l,err := f.Write(data)
	if (err != nil) {return err}

	if (l != len(data)) {return mkErr("unable to write all data")}
	f.Sync()
	return err
}

func readFile(path string,file string) (data []byte, err error) {
	p := path+"/"+file
	data, err = ioutil.ReadFile(p)
	return data,err
}

func mkErr(err string) error {
	return errors.New("holochain: "+err)
}

func dirExists(name string) bool {
	info, err := os.Stat(name)
	return err == nil &&  info.Mode().IsDir();
}

func fileExists(name string) bool {
	info, err := os.Stat(name)
	return err == nil && info.Mode().IsRegular();
}
