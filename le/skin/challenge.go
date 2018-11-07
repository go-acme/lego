package skin

import "github.com/xenolf/lego/le"

type ChallengeService service

func (c *ChallengeService) New(chlgURL string) (le.ChallengeExtend, error) {
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
