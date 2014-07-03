package datasource

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/coreos/coreos-cloudinit/pkg"
)

// metadataService retrieves metadata and userdata from an EC2[1] (2009-04-04)
// compatible endpoint. It attempts to retrieve metadata bit-by-bit from the
// EC2 endpoint, and populates that into an equivalent JSON blob.
//
// [1] http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/AESDG-chapter-instancedata.html#instancedata-data-categories

const (
	BaseUrl     = "http://169.254.169.254/"
	ApiVersion  = "2009-04-04"
	UserdataUrl = BaseUrl + ApiVersion + "/user-data"
	MetadataUrl = BaseUrl + ApiVersion + "/meta-data"
)

type metadataService struct{}

type getter interface {
	GetRetry(string) ([]byte, error)
}

func NewMetadataService() *metadataService {
	return &metadataService{}
}

func (ms *metadataService) IsAvailable() bool {
	client := pkg.NewHttpClient()
	_, err := client.Get(BaseUrl)
	return (err == nil)
}

func (ms *metadataService) AvailabilityChanges() bool {
	return true
}

func (ms *metadataService) ConfigRoot() string {
	return ""
}

func (ms *metadataService) FetchMetadata() ([]byte, error) {
	return fetchMetadata(pkg.NewHttpClient())
}

func (ms *metadataService) FetchUserdata() ([]byte, error) {
	client := pkg.NewHttpClient()
	if data, err := client.GetRetry(UserdataUrl); err == nil {
		return data, err
	} else if _, ok := err.(pkg.ErrTimeout); ok {
		return data, err
	}

	if data, err := client.GetRetry(UserdataUrl); err == nil {
		return data, err
	} else if _, ok := err.(pkg.ErrNotFound); ok {
		return []byte{}, nil
	} else {
		return data, err
	}
}

func (ms *metadataService) Type() string {
	return "metadata-service"
}

func fetchMetadata(client getter) ([]byte, error) {
	attrs := make(map[string]interface{})
	if keynames, err := fetchAttributes(client, fmt.Sprintf("%s/public-keys", MetadataUrl)); err == nil {
		keyIDs := make(map[string]string)
		for _, keyname := range keynames {
			tokens := strings.SplitN(keyname, "=", 2)
			if len(tokens) != 2 {
				return nil, fmt.Errorf("malformed public key: %q\n", keyname)
			}
			keyIDs[tokens[1]] = tokens[0]
		}

		keys := make(map[string]string)
		for name, id := range keyIDs {
			sshkey, err := fetchAttribute(client, fmt.Sprintf("%s/public-keys/%s/openssh-key", MetadataUrl, id))
			if err != nil {
				return nil, err
			}
			keys[name] = sshkey
			fmt.Printf("Found SSH key for %q\n", name)
		}
		attrs["public_keys"] = keys
	} else if _, ok := err.(pkg.ErrNotFound); !ok {
		return nil, err
	}

	if hostname, err := fetchAttribute(client, fmt.Sprintf("%s/hostname", MetadataUrl)); err == nil {
		attrs["hostname"] = hostname
	} else if _, ok := err.(pkg.ErrNotFound); !ok {
		return nil, err
	}

	if localAddr, err := fetchAttribute(client, fmt.Sprintf("%s/local-ipv4", MetadataUrl)); err == nil {
		attrs["local-ipv4"] = localAddr
	} else if _, ok := err.(pkg.ErrNotFound); !ok {
		return nil, err
	}

	if publicAddr, err := fetchAttribute(client, fmt.Sprintf("%s/public-ipv4", MetadataUrl)); err == nil {
		attrs["public-ipv4"] = publicAddr
	} else if _, ok := err.(pkg.ErrNotFound); !ok {
		return nil, err
	}

	if content_path, err := fetchAttribute(client, fmt.Sprintf("%s/network_config/content_path", MetadataUrl)); err == nil {
		attrs["network_config"] = map[string]string{
			"content_path": content_path,
		}
	} else if _, ok := err.(pkg.ErrNotFound); !ok {
		return nil, err
	}

	return json.Marshal(attrs)
}

func fetchAttributes(client getter, url string) ([]string, error) {
	resp, err := client.GetRetry(url)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(bytes.NewBuffer(resp))
	data := make([]string, 0)
	for scanner.Scan() {
		data = append(data, scanner.Text())
	}
	return data, scanner.Err()
}

func fetchAttribute(client getter, url string) (string, error) {
	if attrs, err := fetchAttributes(client, url); err == nil && len(attrs) > 0 {
		return attrs[0], nil
	} else {
		return "", err
	}
}
