// Utility functions for base64url encoding/decoding
function bufferToBase64Url(buffer) {
  return btoa(String.fromCharCode(...new Uint8Array(buffer)))
    .replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '');
}

function base64UrlToBuffer(base64url) {
  const base64 = base64url.replace(/-/g, '+').replace(/_/g, '/');
  const str = atob(base64);
  const buffer = new ArrayBuffer(str.length);
  const byteView = new Uint8Array(buffer);
  for (let i = 0; i < str.length; i++) {
    byteView[i] = str.charCodeAt(i);
  }
  return buffer;
}

class WebAuthnService {
  constructor() {
    this.authToken = null;
  }

  setAuthToken(token) {
    this.authToken = token;
  }

  async getRegistrationOptions() {
    const response = await fetch('/api/user/webauthn/register/options', {
      headers: {
        'Authorization': this.authToken
      },
      credentials: 'include',
      mode: 'cors'
    });

    const rawResponse = await response.text();
    let options;
    try {
      options = JSON.parse(rawResponse);
    } catch (e) {
      throw new Error('Failed to parse server response as JSON: ' + e.message);
    }

    if (!options) {
      throw new Error('Server returned empty options');
    }

    if (!options.publicKey) {
      throw new Error('Server response missing publicKey object');
    }

    options = options.publicKey;

    if (!options.challenge) {
      throw new Error('Server response missing challenge');
    }

    if (!options.rp) {
      throw new Error('Server response missing rp (relying party) information');
    }

    if (!options.user) {
      throw new Error('Server response missing user information');
    }

    // Convert base64url challenge to ArrayBuffer
    options.challenge = base64UrlToBuffer(options.challenge);

    // Convert user.id from base64url to ArrayBuffer if it exists
    if (options.user && options.user.id) {
      options.user.id = base64UrlToBuffer(options.user.id);
    }

    // Convert any existing credentials to ArrayBuffer
    if (options.excludeCredentials) {
      options.excludeCredentials = options.excludeCredentials.map(cred => ({
        ...cred,
        id: base64UrlToBuffer(cred.id)
      }));
    }

    return options;
  }

  async register() {
    if (!this.authToken) {
      throw new Error('Please login first');
    }

    const options = await this.getRegistrationOptions();
    const credential = await navigator.credentials.create({ publicKey: options });

    const credentialData = {
      id: credential.id,
      rawId: bufferToBase64Url(credential.rawId),
      type: credential.type,
      response: {
        attestationObject: bufferToBase64Url(credential.response.attestationObject),
        clientDataJSON: bufferToBase64Url(credential.response.clientDataJSON)
      }
    };

    const response = await fetch('/api/user/webauthn/register', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': this.authToken
      },
      credentials: 'include',
      mode: 'cors',
      body: JSON.stringify(credentialData)
    });

    return await response.text();
  }

  async getLoginOptions(email) {
    const response = await fetch(`/api/user/webauthn/login/options?email=${encodeURIComponent(email)}`, {
      credentials: 'include',
      mode: 'cors'
    });

    const rawResponse = await response.text();
    let options;
    try {
      options = JSON.parse(rawResponse);
    } catch (e) {
      throw new Error('Failed to parse server response as JSON: ' + e.message);
    }

    if (!options) {
      throw new Error('Server returned empty options');
    }

    if (!options.publicKey) {
      throw new Error('Server response missing publicKey object');
    }

    options = options.publicKey;

    if (!options.challenge) {
      throw new Error('Server response missing challenge');
    }

    // Convert base64url challenge to ArrayBuffer
    options.challenge = base64UrlToBuffer(options.challenge);

    // Convert any existing credentials to ArrayBuffer
    if (options.allowCredentials) {
      options.allowCredentials = options.allowCredentials.map(cred => ({
        ...cred,
        id: base64UrlToBuffer(cred.id)
      }));
    }

    return options;
  }

  async login(email) {
    if (!email) {
      throw new Error('Please enter your email');
    }

    const options = await this.getLoginOptions(email);
    const assertion = await navigator.credentials.get({ publicKey: options });

    const assertionData = {
      id: assertion.id,
      rawId: bufferToBase64Url(assertion.rawId),
      type: assertion.type,
      response: {
        authenticatorData: bufferToBase64Url(assertion.response.authenticatorData),
        clientDataJSON: bufferToBase64Url(assertion.response.clientDataJSON),
        signature: bufferToBase64Url(assertion.response.signature),
        userHandle: assertion.response.userHandle ? bufferToBase64Url(assertion.response.userHandle) : null
      }
    };

    const response = await fetch(`/api/user/webauthn/login?email=${encodeURIComponent(email)}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      credentials: 'include',
      mode: 'cors',
      body: JSON.stringify(assertionData)
    });

    return await response.text();
  }
}

// Export the service
window.WebAuthnService = WebAuthnService; 