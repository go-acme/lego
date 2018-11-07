package api

import "github.com/xenolf/lego/le"

type ChallengeService service

func (c *ChallengeService) New(chlgURL string) (le.ChallengeExtend, error) {
	// Challenge initiation is done by sending a JWS payload containing the trivial JSON object `{}`.
	// We use an empty struct instance as the postJSON payload here to achieve this result.
	var chlng le.ChallengeExtend
	resp, err := c.core.post(chlgURL, struct{}{}, &chlng)
	if err != nil {
		return le.ChallengeExtend{}, err
	}

	chlng.AuthorizationURL = getLink(resp.Header, "up")
	chlng.RetryAfter = getRetryAfter(resp)
	return chlng, nil
}

func (c *ChallengeService) Get(chlgURL string) (le.ChallengeExtend, error) {
	var chlng le.ChallengeExtend
	resp, err := c.core.postAsGet(chlgURL, &chlng)
	if err != nil {
		return le.ChallengeExtend{}, err
	}

	chlng.AuthorizationURL = getLink(resp.Header, "up")
	chlng.RetryAfter = getRetryAfter(resp)
	return chlng, nil
}
