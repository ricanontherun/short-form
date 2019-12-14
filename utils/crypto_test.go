package utils

import "testing"

func TestEncryptDecrypt(t *testing.T) {
	e1 := MakeEncryptor("secret")
	e2 := MakeEncryptor("another secret")

	testString := "testing"

	if encryptedBytes, err := e1.Encrypt([]byte(testString)); err != nil {
		t.Error(err.Error())
		t.FailNow()
	} else {
		if string(encryptedBytes) == testString {
			t.Error("Encrypted encryptedBytes should be different than original string")
		} else {
			if decryptedBytes, err := e1.Decrypt(encryptedBytes); err != nil {
				t.Error(err.Error())
				t.FailNow()
			} else {
				decryptedString := string(decryptedBytes)
				if decryptedString != testString {
					t.Errorf("%s != %s", decryptedString, testString)
				}
			}

			// An encryptor initialized with another secret shouldn't be capable
			// to decrypt another thing.
			if _, err := e2.Decrypt(encryptedBytes); err == nil {
				t.Error("Should not have been able to decrypt with another secret")
				t.FailNow()
			}
		}
	}
}
