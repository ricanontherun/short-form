package utils

import "testing"

func TestEncryptDecrypt(t *testing.T) {
	s1Encryptor := MakeEncryptor("secret")
	s2Encryptor := MakeEncryptor("another secret")

	testString := "testing"

	if s1EncryptedBytes, err := s1Encryptor.Encrypt([]byte(testString)); err != nil {
		t.Error(err.Error())
		t.FailNow()
	} else {
		if string(s1EncryptedBytes) == testString {
			t.Error("Encrypted bytes should be different than original string")
		} else {
			if s1DecryptedBytes, err := s1Encryptor.Decrypt(s1EncryptedBytes); err != nil {
				t.Error(err.Error())
				t.FailNow()
			} else {
				decryptedString := string(s1DecryptedBytes)
				if decryptedString != testString {
					t.Errorf("%s != %s", decryptedString, testString)
				}
			}

			// An encryptor initialized with another secret shouldn't be capable
			// to decrypt another thing.
			if _, err := s2Encryptor.Decrypt(s1EncryptedBytes); err == nil {
				t.Error("Should not have been able to decrypt with another secret")
				t.FailNow()
			}
		}
	}
}
