package validation

import (
	"testing"
)

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{"Valid email", "user@example.com", false},
		{"Valid email with subdomain", "user@mail.example.com", false},
		{"Valid email with plus", "user+tag@example.com", false},
		{"Empty email", "", true},
		{"Missing @", "userexample.com", true},
		{"Missing domain", "user@", true},
		{"Missing local part", "@example.com", true},
		{"Invalid characters", "user name@example.com", true},
		{"Too long", "verylongemailaddressthatexceedsthemaximumlengthallowedforemailaddressesaccordingtorfc5321specifications@verylongdomainnamethatexceedsthemaximumlengthallowedfordomainnamesaccordingtorfc1035specifications.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"Valid strong password", "MyP@ssw0rd!", false},
		{"Valid complex password", "C0mpl3x!Pass", false},
		{"Too short", "Sh0rt!", true},
		{"No uppercase", "myp@ssw0rd!", true},
		{"No lowercase", "MYP@SSW0RD!", true},
		{"No number", "MyPassword!", true},
		{"No special char", "MyPassword0", true},
		{"Common password", "password", true},
		{"Common password 2", "password123", true},
		{"Repeated chars", "Aaaa1234!", true},
		{"Sequential chars", "Abcd1234!", true},
		{"Too long", "MyP@ssw0rd!" + string(make([]byte, 120)), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePasswordWithRules(t *testing.T) {
	// Test custom rules (less strict)
	relaxedRules := PasswordComplexity{
		MinLength:      6,
		RequireUpper:   false,
		RequireLower:   true,
		RequireNumber:  true,
		RequireSpecial: false,
	}

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"Valid relaxed password", "simple123", false},
		{"Too short", "abc12", true},
		{"No lowercase", "SIMPLE123", true},
		{"No number", "simplepass", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePasswordWithRules(tt.password, relaxedRules)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePasswordWithRules() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Normal string", "hello world", "hello world"},
		{"With null bytes", "hello\x00world", "helloworld"},
		{"With whitespace", "  hello world  ", "hello world"},
		{"Multiple null bytes", "a\x00b\x00c", "abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeString(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeString() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSanitizeHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Normal text", "hello world", "hello world"},
		{"With tags", "<script>alert('xss')</script>", "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;"},
		{"With quotes", `Hello "World"`, "Hello &quot;World&quot;"},
		{"With ampersand", "Tom & Jerry", "Tom &amp; Jerry"},
		{"Multiple special chars", `<div class="test">Data & "stuff"</div>`, "&lt;div class=&quot;test&quot;&gt;Data &amp; &quot;stuff&quot;&lt;/div&gt;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeHTML(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeHTML() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		wantErr  bool
	}{
		{"Valid username", "john_doe", false},
		{"Valid with numbers", "user123", false},
		{"Valid with hyphen", "john-doe", false},
		{"Too short", "ab", true},
		{"Too long", "thisusernameiswaytoolongtobevalid123", true},
		{"Empty", "", true},
		{"With spaces", "john doe", true},
		{"With special chars", "john@doe", true},
		{"Starts with underscore", "_johndoe", true},
		{"Ends with hyphen", "johndoe-", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUsername(tt.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUsername() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"Valid HTTP URL", "http://example.com", false},
		{"Valid HTTPS URL", "https://example.com", false},
		{"Valid with path", "https://example.com/path/to/page", false},
		{"Valid with query", "https://example.com?param=value", false},
		{"Empty URL", "", true},
		{"Invalid protocol", "ftp://example.com", true},
		{"Missing protocol", "example.com", true},
		{"Invalid format", "not a url", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHasRepeatedChars(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		n        int
		expected bool
	}{
		{"No repeats", "abcdef", 3, false},
		{"Has repeats", "aaabbb", 3, true},
		{"Exactly n repeats", "aaa", 3, true},
		{"Less than n repeats", "aa", 3, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasRepeatedChars(tt.input, tt.n)
			if result != tt.expected {
				t.Errorf("hasRepeatedChars() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestHasSequentialChars(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		n        int
		expected bool
	}{
		{"No sequence", "aceg", 3, false},
		{"Has ascending sequence", "abcd", 3, true},
		{"Has descending sequence", "dcba", 3, true},
		{"Exactly n sequential", "abc", 3, true},
		{"Less than n sequential", "ab", 3, false},
		{"Numbers sequential", "1234", 3, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasSequentialChars(tt.input, tt.n)
			if result != tt.expected {
				t.Errorf("hasSequentialChars() = %v, want %v", result, tt.expected)
			}
		})
	}
}
