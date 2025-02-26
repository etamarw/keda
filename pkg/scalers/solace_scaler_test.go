package scalers

import (
	"fmt"
	"net/http"
	"testing"

	"k8s.io/api/autoscaling/v2beta2"
)

type testSolaceMetadata struct {
	testID   string
	metadata map[string]string
	isError  bool
}

var (
	soltestValidBaseURL        = "http://localhost:8080"
	soltestValidUsername       = "admin"
	soltestValidPassword       = "admin"
	soltestValidVpn            = "dennis_vpn"
	soltestValidQueueName      = "queue3"
	soltestValidMsgCountTarget = "10"
	soltestValidMsgSpoolTarget = "20"
	soltestEnvUsername         = "SOLTEST_USERNAME"
	soltestEnvPassword         = "SOLTEST_PASSWORD"
)

// AUTH RECORD FOR TEST
var testDataSolaceAuthParamsVALID = map[string]string{
	solaceMetaUsername: soltestValidUsername,
	solaceMetaPassword: soltestValidPassword,
}

// ENV VARS FOR TEST -- VALID USER / PWD
var testDataSolaceResolvedEnvVALID = map[string]string{
	soltestEnvUsername: soltestValidUsername, // Sets the environment variables to the correct values
	soltestEnvPassword: soltestValidPassword,
}

// TEST CASES FOR SolaceParseMetadata()
var testParseSolaceMetadata = []testSolaceMetadata{
	// Empty
	{
		"#001 - EMPTY", map[string]string{},
		true,
	},
	// +Case - brokerBaseUrl
	{
		"#002 - brokerBaseUrl",
		map[string]string{
			"":                        "",
			solaceMetaSempBaseURL:     soltestValidBaseURL,
			solaceMetaMsgVpn:          soltestValidVpn,
			solaceMetaUsernameFromEnv: "",
			solaceMetaPasswordFromEnv: "",
			solaceMetaUsername:        soltestValidUsername,
			solaceMetaPassword:        soltestValidPassword,
			solaceMetaQueueName:       soltestValidQueueName,
			solaceMetaMsgCountTarget:  soltestValidMsgCountTarget,
		},
		false,
	},
	// -Case - missing username (clear)
	{
		"#007 - missing username (clear)",
		map[string]string{
			solaceMetaSempBaseURL:     soltestValidBaseURL,
			solaceMetaMsgVpn:          soltestValidVpn,
			solaceMetaUsernameFromEnv: "",
			solaceMetaPasswordFromEnv: "",
			solaceMetaUsername:        "",
			solaceMetaPassword:        soltestValidPassword,
			solaceMetaQueueName:       soltestValidQueueName,
			solaceMetaMsgCountTarget:  soltestValidMsgCountTarget,
		},
		true,
	},
	// -Case - missing password (clear)
	{
		"#008 - missing password (clear)",
		map[string]string{
			solaceMetaSempBaseURL:     soltestValidBaseURL,
			solaceMetaMsgVpn:          soltestValidVpn,
			solaceMetaUsernameFromEnv: "",
			solaceMetaPasswordFromEnv: "",
			solaceMetaUsername:        soltestValidUsername,
			solaceMetaPassword:        "",
			solaceMetaQueueName:       soltestValidQueueName,
			solaceMetaMsgCountTarget:  soltestValidMsgCountTarget,
		},
		true,
	},
	// -Case - missing queue
	{
		"#009 - missing queueName",
		map[string]string{
			solaceMetaSempBaseURL:     soltestValidBaseURL,
			solaceMetaMsgVpn:          soltestValidVpn,
			solaceMetaUsernameFromEnv: "",
			solaceMetaPasswordFromEnv: "",
			solaceMetaUsername:        soltestValidUsername,
			solaceMetaPassword:        soltestValidPassword,
			solaceMetaQueueName:       "",
			solaceMetaMsgCountTarget:  soltestValidMsgCountTarget,
		},
		true,
	},
	// -Case - missing msgCountTarget
	{
		"#010 - missing msgCountTarget",
		map[string]string{
			solaceMetaSempBaseURL:         soltestValidBaseURL,
			solaceMetaMsgVpn:              soltestValidVpn,
			solaceMetaUsernameFromEnv:     "",
			solaceMetaPasswordFromEnv:     "",
			solaceMetaUsername:            soltestValidUsername,
			solaceMetaPassword:            soltestValidPassword,
			solaceMetaQueueName:           soltestValidQueueName,
			solaceMetaMsgCountTarget:      "",
			solaceMetaMsgSpoolUsageTarget: "",
		},
		true,
	},
	// -Case - msgSpoolUsageTarget non-numeric
	{
		"#011 - msgSpoolUsageTarget non-numeric",
		map[string]string{
			solaceMetaSempBaseURL:     soltestValidBaseURL,
			solaceMetaMsgVpn:          soltestValidVpn,
			solaceMetaUsernameFromEnv: "",
			solaceMetaPasswordFromEnv: "",
			solaceMetaUsername:        soltestValidUsername,
			solaceMetaPassword:        soltestValidPassword,
			solaceMetaQueueName:       soltestValidQueueName,
			solaceMetaMsgCountTarget:  "NOT_AN_INTEGER",
		},
		true,
	},
	// -Case - msgSpoolUsage non-numeric
	{
		"#012 - msgSpoolUsage non-numeric",
		map[string]string{
			solaceMetaSempBaseURL:         soltestValidBaseURL,
			solaceMetaMsgVpn:              soltestValidVpn,
			solaceMetaUsernameFromEnv:     "",
			solaceMetaPasswordFromEnv:     "",
			solaceMetaUsername:            soltestValidUsername,
			solaceMetaPassword:            soltestValidPassword,
			solaceMetaQueueName:           soltestValidQueueName,
			solaceMetaMsgSpoolUsageTarget: "NOT_AN_INTEGER",
		},
		true,
	},
	// +Case - Pass with msgSpoolUsageTarget and not msgCountTarget
	{
		"#013 - brokerBaseUrl",
		map[string]string{
			solaceMetaSempBaseURL:         soltestValidBaseURL,
			solaceMetaMsgVpn:              soltestValidVpn,
			solaceMetaUsernameFromEnv:     "",
			solaceMetaPasswordFromEnv:     "",
			solaceMetaUsername:            soltestValidUsername,
			solaceMetaPassword:            soltestValidPassword,
			solaceMetaQueueName:           soltestValidQueueName,
			solaceMetaMsgSpoolUsageTarget: soltestValidMsgSpoolTarget,
		},
		false,
	},
}

var testSolaceEnvCreds = []testSolaceMetadata{
	// +Case - Should find ENV vars
	{
		"#101 - Connect with Credentials in env",
		map[string]string{
			solaceMetaSempBaseURL:     soltestValidBaseURL,
			solaceMetaMsgVpn:          soltestValidVpn,
			solaceMetaUsernameFromEnv: soltestEnvUsername,
			solaceMetaPasswordFromEnv: soltestEnvPassword,
			//		solaceMetaUsername:              "",
			//		solaceMetaPassword:              "",
			solaceMetaQueueName:      soltestValidQueueName,
			solaceMetaMsgCountTarget: soltestValidMsgCountTarget,
		},
		false,
	},
	// -Case - Should fail with ENV var not found
	{
		"#102 - Environment vars referenced but not found",
		map[string]string{
			solaceMetaSempBaseURL:     soltestValidBaseURL,
			solaceMetaMsgVpn:          soltestValidVpn,
			solaceMetaUsernameFromEnv: "SOLTEST_DNE",
			solaceMetaPasswordFromEnv: "SOLTEST_DNE",
			//		solaceMetaUsername:              "",
			//		solaceMetaPassword:              "",
			solaceMetaQueueName:      soltestValidQueueName,
			solaceMetaMsgCountTarget: soltestValidMsgCountTarget,
		},
		true,
	},
}

var testSolaceK8sSecretCreds = []testSolaceMetadata{
	// Records require Auth Record to be passed

	// +Case - Should find
	{
		"#201 - Connect with credentials from Auth Record (ENV VAR Present)",
		map[string]string{
			solaceMetaSempBaseURL:     soltestValidBaseURL,
			solaceMetaMsgVpn:          soltestValidVpn,
			solaceMetaUsernameFromEnv: soltestEnvUsername,
			solaceMetaPasswordFromEnv: soltestEnvPassword,
			//		solaceMetaUsername:              "",
			//		solaceMetaPassword:              "",
			solaceMetaQueueName:      soltestValidQueueName,
			solaceMetaMsgCountTarget: soltestValidMsgCountTarget,
		},
		false,
	},
	// +Case - should find creds
	{
		"#202 - Connect with credentials from Auth Record (ENV VAR and Clear Auth not present)",
		map[string]string{
			solaceMetaSempBaseURL: soltestValidBaseURL,
			solaceMetaMsgVpn:      soltestValidVpn,
			//		solaceMetaUsernameFromEnv:    soltestEnvUsername,
			//		solaceMetaPasswordFromEnv:    soltestEnvPassword,
			//		solaceMetaUsername:              "",
			//		solaceMetaPassword:              "",
			solaceMetaQueueName:      soltestValidQueueName,
			solaceMetaMsgCountTarget: soltestValidMsgCountTarget,
		},
		false,
	},
	// +Case - Should find with creds
	{
		"#203 - Connect with credentials from Auth Record (ENV VAR Present, Clear Auth not present)",
		map[string]string{
			solaceMetaSempBaseURL:     soltestValidBaseURL,
			solaceMetaMsgVpn:          soltestValidVpn,
			solaceMetaUsernameFromEnv: "SOLTEST_DNE",
			solaceMetaPasswordFromEnv: "SOLTEST_DNE",
			//		solaceMetaUsername:              "",
			//		solaceMetaPassword:              "",
			solaceMetaQueueName:      soltestValidQueueName,
			solaceMetaMsgCountTarget: soltestValidMsgCountTarget,
		},
		false,
	},
}

var testSolaceGetMetricSpecData = []testSolaceMetadata{
	{
		"#401 - Get Metric Spec - msgCountTarget",
		map[string]string{
			solaceMetaSempBaseURL:     soltestValidBaseURL,
			solaceMetaMsgVpn:          soltestValidVpn,
			solaceMetaUsernameFromEnv: "",
			solaceMetaPasswordFromEnv: "",
			solaceMetaUsername:        soltestValidUsername,
			solaceMetaPassword:        soltestValidPassword,
			solaceMetaQueueName:       soltestValidQueueName,
			solaceMetaMsgCountTarget:  soltestValidMsgCountTarget,
			//			solaceMetaMsgSpoolUsageTarget: soltestValidMsgSpoolTarget,
		},
		false,
	},
	{
		"#402 - Get Metric Spec - msgSpoolUsageTarget",
		map[string]string{
			solaceMetaSempBaseURL:     soltestValidBaseURL,
			solaceMetaMsgVpn:          soltestValidVpn,
			solaceMetaUsernameFromEnv: "",
			solaceMetaPasswordFromEnv: "",
			solaceMetaUsername:        soltestValidUsername,
			solaceMetaPassword:        soltestValidPassword,
			solaceMetaQueueName:       soltestValidQueueName,
			//			solaceMetaMsgCountTarget:      soltestValidMsgCountTarget,
			solaceMetaMsgSpoolUsageTarget: soltestValidMsgSpoolTarget,
		},
		false,
	},
	{
		"#403 - Get Metric Spec - BOTH msgSpoolUsage and msgCountTarget",
		map[string]string{
			solaceMetaSempBaseURL:         soltestValidBaseURL,
			solaceMetaMsgVpn:              soltestValidVpn,
			solaceMetaUsernameFromEnv:     "",
			solaceMetaPasswordFromEnv:     "",
			solaceMetaUsername:            soltestValidUsername,
			solaceMetaPassword:            soltestValidPassword,
			solaceMetaQueueName:           soltestValidQueueName,
			solaceMetaMsgCountTarget:      soltestValidMsgCountTarget,
			solaceMetaMsgSpoolUsageTarget: soltestValidMsgSpoolTarget,
		},
		false,
	},
	{
		"#404 - Get Metric Spec - BOTH MISSING",
		map[string]string{
			solaceMetaSempBaseURL:     soltestValidBaseURL,
			solaceMetaMsgVpn:          soltestValidVpn,
			solaceMetaUsernameFromEnv: "",
			solaceMetaPasswordFromEnv: "",
			solaceMetaUsername:        soltestValidUsername,
			solaceMetaPassword:        soltestValidPassword,
			solaceMetaQueueName:       soltestValidQueueName,
			//			solaceMetaMsgCountTarget:      soltestValidMsgCountTarget,
			//			solaceMetaMsgSpoolUsageTarget: soltestValidMsgSpoolTarget,
		},
		true,
	},
	{
		"#405 - Get Metric Spec - BOTH ZERO",
		map[string]string{
			solaceMetaSempBaseURL:         soltestValidBaseURL,
			solaceMetaMsgVpn:              soltestValidVpn,
			solaceMetaUsernameFromEnv:     "",
			solaceMetaPasswordFromEnv:     "",
			solaceMetaUsername:            soltestValidUsername,
			solaceMetaPassword:            soltestValidPassword,
			solaceMetaQueueName:           soltestValidQueueName,
			solaceMetaMsgCountTarget:      "0",
			solaceMetaMsgSpoolUsageTarget: "0",
		},
		true,
	},
	{
		"#406 - Get Metric Spec - ONE ZERO; OTHER VALID",
		map[string]string{
			solaceMetaSempBaseURL:         soltestValidBaseURL,
			solaceMetaMsgVpn:              soltestValidVpn,
			solaceMetaUsernameFromEnv:     "",
			solaceMetaPasswordFromEnv:     "",
			solaceMetaUsername:            soltestValidUsername,
			solaceMetaPassword:            soltestValidPassword,
			solaceMetaQueueName:           soltestValidQueueName,
			solaceMetaMsgCountTarget:      "0",
			solaceMetaMsgSpoolUsageTarget: soltestValidMsgSpoolTarget,
		},
		false,
	},
}

var testSolaceExpectedMetricNames = map[string]string{
	solaceScalerID + "-" + soltestValidVpn + "-" + soltestValidQueueName + "-" + solaceTriggermsgcount:      "",
	solaceScalerID + "-" + soltestValidVpn + "-" + soltestValidQueueName + "-" + solaceTriggermsgspoolusage: "",
}

func TestSolaceParseSolaceMetadata(t *testing.T) {
	for _, testData := range testParseSolaceMetadata {
		fmt.Print(testData.testID)
		_, err := parseSolaceMetadata(&ScalerConfig{ResolvedEnv: nil, TriggerMetadata: testData.metadata, AuthParams: nil})
		switch {
		case err != nil && !testData.isError:
			t.Error("expected success but got error: ", err)
			fmt.Println(" --> FAIL")
		case testData.isError && err == nil:
			t.Error("expected error but got success")
			fmt.Println(" --> FAIL")
		default:
			fmt.Println(" --> PASS")
		}
	}
	for _, testData := range testSolaceEnvCreds {
		fmt.Print(testData.testID)
		_, err := parseSolaceMetadata(&ScalerConfig{ResolvedEnv: testDataSolaceResolvedEnvVALID, TriggerMetadata: testData.metadata, AuthParams: nil})
		switch {
		case err != nil && !testData.isError:
			t.Error("expected success but got error: ", err)
			fmt.Println(" --> FAIL")
		case testData.isError && err == nil:
			t.Error("expected error but got success")
			fmt.Println(" --> FAIL")
		default:
			fmt.Println(" --> PASS")
		}
	}
	for _, testData := range testSolaceK8sSecretCreds {
		fmt.Print(testData.testID)
		_, err := parseSolaceMetadata(&ScalerConfig{ResolvedEnv: nil, TriggerMetadata: testData.metadata, AuthParams: testDataSolaceAuthParamsVALID})
		switch {
		case err != nil && !testData.isError:
			t.Error("expected success but got error: ", err)
			fmt.Println(" --> FAIL")
		case testData.isError && err == nil:
			t.Error("expected error but got success")
			fmt.Println(" --> FAIL")
		default:
			fmt.Println(" --> PASS")
		}
	}
}

func TestSolaceGetMetricSpec(t *testing.T) {
	for idx := 0; idx < len(testSolaceGetMetricSpecData); idx++ {
		testData := testSolaceGetMetricSpecData[idx]
		fmt.Print(testData.testID)
		var err error
		var solaceMeta *SolaceMetadata
		solaceMeta, err = parseSolaceMetadata(&ScalerConfig{ResolvedEnv: testDataSolaceResolvedEnvVALID, TriggerMetadata: testData.metadata, AuthParams: testDataSolaceAuthParamsVALID})
		if err != nil {
			fmt.Printf("\n       Failed to parse metadata: %v", err)
		} else {
			// DECLARE SCALER AND RUN METHOD TO GET METRICS
			testSolaceScaler := SolaceScaler{
				metadata:   solaceMeta,
				httpClient: http.DefaultClient,
			}

			var metric []v2beta2.MetricSpec
			if metric = testSolaceScaler.GetMetricSpecForScaling(); len(metric) == 0 {
				err = fmt.Errorf("metric value not found")
			} else {
				metricName := metric[0].External.Metric.Name
				if _, ok := testSolaceExpectedMetricNames[metricName]; ok == false {
					err = fmt.Errorf("expected Metric value not found")
				}
			}
		}
		switch {
		case testData.isError && err == nil:
			fmt.Println(" --> FAIL")
			t.Error("expected to fail but passed", err)
		case !testData.isError && err != nil:
			fmt.Println(" --> FAIL")
			t.Error("expected success but failed", err)
		default:
			fmt.Println(" --> PASS")
		}
	}
}
