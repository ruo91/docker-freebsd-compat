package graph

import (
	"encoding/json"
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/docker/distribution/digest"
	"github.com/docker/docker/registry"
	"github.com/docker/docker/trust"
	"github.com/docker/docker/utils"
	"github.com/docker/libtrust"
)

// loadManifest loads a manifest from a byte array and verifies its content.
// The signature must be verified or an error is returned. If the manifest
// contains no signatures by a trusted key for the name in the manifest, the
// image is not considered verified. The parsed manifest object and a boolean
// for whether the manifest is verified is returned.
func (s *TagStore) loadManifest(manifestBytes []byte, dgst, ref string) (*registry.ManifestData, bool, error) {
	sig, err := libtrust.ParsePrettySignature(manifestBytes, "signatures")
	if err != nil {
		return nil, false, fmt.Errorf("error parsing payload: %s", err)
	}

	keys, err := sig.Verify()
	if err != nil {
		return nil, false, fmt.Errorf("error verifying payload: %s", err)
	}

	payload, err := sig.Payload()
	if err != nil {
		return nil, false, fmt.Errorf("error retrieving payload: %s", err)
	}

	var manifestDigest digest.Digest

	if dgst != "" {
		manifestDigest, err = digest.ParseDigest(dgst)
		if err != nil {
			return nil, false, fmt.Errorf("invalid manifest digest from registry: %s", err)
		}

		dgstVerifier, err := digest.NewDigestVerifier(manifestDigest)
		if err != nil {
			return nil, false, fmt.Errorf("unable to verify manifest digest from registry: %s", err)
		}

		dgstVerifier.Write(payload)

		if !dgstVerifier.Verified() {
			computedDigest, _ := digest.FromBytes(payload)
			return nil, false, fmt.Errorf("unable to verify manifest digest: registry has %q, computed %q", manifestDigest, computedDigest)
		}
	}

	if utils.DigestReference(ref) && ref != manifestDigest.String() {
		return nil, false, fmt.Errorf("mismatching image manifest digest: got %q, expected %q", manifestDigest, ref)
	}

	var manifest registry.ManifestData
	if err := json.Unmarshal(payload, &manifest); err != nil {
		return nil, false, fmt.Errorf("error unmarshalling manifest: %s", err)
	}
	if manifest.SchemaVersion != 1 {
		return nil, false, fmt.Errorf("unsupported schema version: %d", manifest.SchemaVersion)
	}

	var verified bool
	for _, key := range keys {
		namespace := manifest.Name
		if namespace[0] != '/' {
			namespace = "/" + namespace
		}
		b, err := key.MarshalJSON()
		if err != nil {
			return nil, false, fmt.Errorf("error marshalling public key: %s", err)
		}
		// Check key has read/write permission (0x03)
		v, err := s.trustService.CheckKey(namespace, b, 0x03)
		if err != nil {
			vErr, ok := err.(trust.NotVerifiedError)
			if !ok {
				return nil, false, fmt.Errorf("error running key check: %s", err)
			}
			logrus.Debugf("Key check result: %v", vErr)
		}
		verified = v
		if verified {
			logrus.Debug("Key check result: verified")
		}
	}
	return &manifest, verified, nil
}

func checkValidManifest(manifest *registry.ManifestData) error {
	if len(manifest.FSLayers) != len(manifest.History) {
		return fmt.Errorf("length of history not equal to number of layers")
	}

	if len(manifest.FSLayers) == 0 {
		return fmt.Errorf("no FSLayers in manifest")
	}

	return nil
}
