package vault

import (
	log "github.com/Sirupsen/logrus"
)

// WalkNode iterates on a secret path that appears to be folder
func walkNode(v *Client, path string) {
	log.Debug("walk ", path)

	secretsList, err := v.client.Logical().List(path)
	if err != nil {
		log.Fatal(err)
	}

	for a, b := range secretsList.Data {
		log.Debug(a, b)
		for _, p := range b.([]interface{}) {
			if p.(string)[len(p.(string))-1:] == "/" {
				node := p.(string)
				walkNode(v, path+"/"+node)
			} else {
				secret, err := v.ReadSecret(path + "/" + p.(string))
				if err != nil {
					log.Panic(err)
				}
				log.Info(secret)
			}
		}
	}
}
