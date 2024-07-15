// copied from https://gist.github.com/rishabhpoddar/8c26ed237add1a5b86481e72032abf8d
package shield

import (
	"time"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/supertokens/supertokens-golang/recipe/session/sessmodels"
)

func (a *AuthenticatorSuperTokens) getJWKSFromCacheIfPresent() *sessmodels.GetJWKSResult {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	if a.jwksCache != nil {
		// This means that we have valid JWKs for the given core path
		// We check if we need to refresh before returning
		currentTime := time.Now().UnixNano() / int64(time.Millisecond)

		// This means that the value in cache is not expired, in this case we return the cached value
		//
		// Note that this also means that the SDK will not try to query any other Core (if there are multiple)
		// if it has a valid cache entry from one of the core URLs. It will only attempt to fetch
		// from the cores again after the entry in the cache is expired
		if (currentTime - a.jwksCache.LastFetched) < a.jwkCacheMaxAgeMiliseconds {
			return a.jwksCache
		}
	}

	return nil
}

func (a *AuthenticatorSuperTokens) getJWKS() (*keyfunc.JWKS, error) {
	resultFromCache := a.getJWKSFromCacheIfPresent()
	if resultFromCache != nil {
		return resultFromCache.JWKS, nil
	}

	a.mutex.Lock()
	defer a.mutex.Unlock()
	// RefreshUnknownKID - Fetch JWKS again if the kid in the header of the JWT does not match any in
	// the keyfunc library's cache
	jwks, err := keyfunc.Get(a.coreURL, keyfunc.Options{
		RefreshUnknownKID: true,
	})
	if err != nil {
		return nil, err
	}

	jwksResult := sessmodels.GetJWKSResult{
		JWKS:        jwks,
		Error:       nil,
		LastFetched: time.Now().UnixNano() / int64(time.Millisecond),
	}

	a.jwksCache = &jwksResult
	return jwksResult.JWKS, nil
}
