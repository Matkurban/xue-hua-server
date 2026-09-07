package prommetrics

import "testing"

//func TestNewGrpcPromObj(t *testing.T) {
//	// Create a custom metric to pass into the NewGrpcPromObj function.
//	customMetric := prometheus.NewCounter(prometheus.CounterOpts{
//		Name: "test_metric",
//		Help: "This is a test metric.",
//	})
//	cusMetrics := []prometheus.Collector{customMetric}
//
//	// Call NewGrpcPromObj with the custom metrics.
//	reg, grpcMetrics, err := NewGrpcPromObj(cusMetrics)
//
//	// Assert no error was returned.
//	assert.NoError(t, err)
//
//	// Assert the registry was correctly initialized.
//	assert.NotNil(t, reg)
//
//	// Assert the grpcMetrics was correctly initialized.
//	assert.NotNil(t, grpcMetrics)
//
//	// Assert that the custom metric is registered.
//	mfs, err := reg.Gather()
//	assert.NoError(t, err)
//	assert.NotEmpty(t, mfs) // Ensure some metrics are present.
//	found := false
//	for _, mf := range mfs {
//		if *mf.Name == "test_metric" {
//			found = true
//			break
//		}
//	}
//	assert.True(t, found, "Custom metric not found in registry")
//}

//func TestGetGrpcCusMetrics(t *testing.T) {
//	conf := config2.NewGlobalConfig()
//
//	config2.InitConfig(conf, "../../config")
//	// Test various cases based on the switch statement in the GetGrpcCusMetrics function.
//	testCases := []struct {
//		name     string
//		expected int // The expected number of metrics for each case.
//	}{
//		{conf.RpcRegisterName.OpenImMessageGatewayName, 1},
//	}
//
//	for _, tc := range testCases {
//		t.Run(tc.name, func(t *testing.T) {
//			metrics := GetGrpcCusMetrics(tc.name, &conf.RpcRegisterName)
//			assert.Len(t, metrics, tc.expected)
//		})
//	}
//}

func TestName(t *testing.T) {
	RegistryApi()
	RegistryApi()

}
