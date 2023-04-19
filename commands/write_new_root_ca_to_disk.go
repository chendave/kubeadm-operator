/*
Copyright 2023 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package commands

import (
	"os"

	"io/fs"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	operatorv1 "k8s.io/kubeadm/operator/api/v1alpha1"
)

func runWriteNewRootCaToDisk(spec *operatorv1.WriteNewRootCaToDiskSpec, log logr.Logger) error {
	log.Info("Writing new root CA on controller plane - N.")
	_, err := os.Stat(CertDir + NewCaCertFilename)
	if err == nil {
		log.Error(err, CertDir+NewCaCertFilename+" already exists.")
		return err
	}
	if !errors.Is(err, fs.ErrNotExist) {
		log.Error(err, "Error appears while stat:"+CertDir+NewCaCertFilename)
		return err
	}

	// backup {ca.crt, ca.key} to {ca.crt.old, ca.key.old}
	log.Info("Backing up old CA on controller plane - N.")
	err = os.Rename(CertDir+OriginalCaCertFilename, CertDir+OriginalCaCertFilename+".old")
	if err != nil {
		log.Error(err, "Cant's backup:"+CertDir+OriginalCaCertFilename)
		return err
	}
	err = os.Rename(CertDir+OriginalCaKeyFilename, CertDir+OriginalCaKeyFilename+".old")
	if err != nil {
		log.Error(err, "Cant's backup:"+CertDir+OriginalCaKeyFilename)
		return err
	}

	// writing new CA-pair to {ca.crt.new, ca.key}
	log.Info("Writing new CA on controller plane - N.")
	newCaCertFile, err := os.Create(CertDir + NewCaCertFilename)
	if err != nil {
		log.Error(err, "Can't create:"+CertDir+NewCaCertFilename)
		return err
	}
	defer newCaCertFile.Close()
	_, err = newCaCertFile.Write(spec.CaRotationOperation.NewCaCert)
	if err != nil {
		log.Error(err, "Can't write file:"+CertDir+NewCaCertFilename)
	}

	newCaKeyFile, err := os.Create(CertDir + OriginalCaKeyFilename)
	if err != nil {
		log.Error(err, "Can't create:"+CertDir+OriginalCaKeyFilename)
		return err
	}
	defer newCaKeyFile.Close()
	_, err = newCaKeyFile.Write(spec.CaRotationOperation.NewCaKey)
	if err != nil {
		log.Error(err, "Can't write file:"+CertDir+OriginalCaKeyFilename)
	}

	// writing CA bundle to {ca.crt}
	log.Info("Writing CA bundle on controller plane - N ")
	newCaBundleFile, err := os.Create(CertDir + CaBundleFilename)
	if err != nil {
		log.Error(err, "Can't create:"+CertDir+CaBundleFilename)
	}
	defer newCaBundleFile.Close()
	_, err = newCaBundleFile.Write(spec.CaRotationOperation.CaBundle)
	if err != nil {
		log.Error(err, "Can't write file:"+CertDir+CaBundleFilename)
	}

	log.Info("Wrote new root CA on controller plane - N.")
	return nil
}
