package oas

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
)

func TestOAS_BuildDefaultTykExtension(t *testing.T) {
	t.Parallel()

	t.Run("build tyk extension with no supplied params", func(t *testing.T) {
		oasDef := OAS{
			T: openapi3.T{
				Info: &openapi3.Info{
					Title: "OAS API",
				},
				Servers: openapi3.Servers{
					{
						URL: "https://example-org.com/api",
					},
				},
			},
		}

		err := oasDef.BuildDefaultTykExtension(TykExtensionConfigParams{})
		assert.NoError(t, err)

		expectedTykExtension := XTykAPIGateway{
			Server: Server{
				ListenPath: ListenPath{
					Value: "/",
				},
			},
			Upstream: Upstream{
				URL: "https://example-org.com/api",
			},
			Info: Info{
				Name: "OAS API",
				State: State{
					Active: true,
				},
			},
		}

		assert.Equal(t, expectedTykExtension, *oasDef.GetTykExtension())
	})

	t.Run("build tyk extension with supplied params", func(t *testing.T) {
		oasDef := OAS{
			T: openapi3.T{
				Info: &openapi3.Info{
					Title: "OAS API",
				},
				Servers: openapi3.Servers{
					{
						URL: "https://example-org.com/api",
					},
				},
			},
		}

		customDomain := "custom-domain.org"
		err := oasDef.BuildDefaultTykExtension(TykExtensionConfigParams{
			ListenPath:   "/listen-api",
			UpstreamURL:  "https://example.org/api",
			CustomDomain: customDomain,
		})

		assert.Nil(t, err)

		expectedTykExtension := XTykAPIGateway{
			Server: Server{
				ListenPath: ListenPath{
					Value: "/listen-api",
				},
				CustomDomain: customDomain,
			},
			Upstream: Upstream{
				URL: "https://example.org/api",
			},
			Info: Info{
				Name: "OAS API",
				State: State{
					Active: true,
				},
			},
		}

		assert.Equal(t, expectedTykExtension, *oasDef.GetTykExtension())
	})

	t.Run("do not override existing tyk extension by default", func(t *testing.T) {
		oasDef := OAS{
			T: openapi3.T{
				Info: &openapi3.Info{
					Title: "OAS API",
				},
				Servers: openapi3.Servers{
					{
						URL: "https://example-org.com/api",
					},
				},
			},
		}

		existingTykExtension := XTykAPIGateway{
			Info: Info{
				Name: "New OAS API",
			},
			Server: Server{
				ListenPath: ListenPath{
					Value: "/new-listen-path",
				},
			},
		}

		oasDef.SetTykExtension(&existingTykExtension)

		err := oasDef.BuildDefaultTykExtension(TykExtensionConfigParams{})
		assert.Nil(t, err)

		expectedTykExtension := XTykAPIGateway{
			Server: Server{
				ListenPath: ListenPath{
					Value: "/new-listen-path",
				},
			},
			Upstream: Upstream{
				URL: "https://example-org.com/api",
			},
			Info: Info{
				Name: "New OAS API",
				State: State{
					Active: true,
				},
			},
		}

		assert.Equal(t, expectedTykExtension, *oasDef.GetTykExtension())
	})

	t.Run("override existing tyk extension with supplied params", func(t *testing.T) {
		const (
			testSSMyAuth = "my_auth"
			testHeader   = "my-header"
		)
		oasDef := OAS{
			T: openapi3.T{
				Info: &openapi3.Info{
					Title: "OAS API",
				},
				Security: openapi3.SecurityRequirements{
					{testSSMyAuth: []string{}},
				},
				Components: openapi3.Components{
					SecuritySchemes: openapi3.SecuritySchemes{
						testSSMyAuth: &openapi3.SecuritySchemeRef{
							Value: openapi3.NewSecurityScheme().WithType(typeApiKey).WithIn(header).WithName(testHeader),
						},
					},
				},
				Servers: openapi3.Servers{
					{
						URL: "https://example-org.com/api",
					},
				},
			},
		}

		existingTykExtension := XTykAPIGateway{
			Info: Info{
				Name: "New OAS API",
			},
			Server: Server{
				ListenPath: ListenPath{
					Value: "/listen-api",
				},
				CustomDomain: "custom-domain.org",
			},
		}

		oasDef.SetTykExtension(&existingTykExtension)

		newCustomDomain := "new-custom-domain.org"

		err := oasDef.BuildDefaultTykExtension(TykExtensionConfigParams{
			ListenPath:     "/new-listen-api",
			UpstreamURL:    "https://example.org/api",
			Authentication: getBoolPointer(true),
			CustomDomain:   newCustomDomain,
		})

		assert.Nil(t, err)

		expectedTykExtension := XTykAPIGateway{
			Server: Server{
				ListenPath: ListenPath{
					Value: "/new-listen-api",
				},
				CustomDomain: newCustomDomain,
				Authentication: &Authentication{
					Enabled: true,
					SecuritySchemes: SecuritySchemes{
						testSSMyAuth: &Token{
							Enabled: true,
							AuthSources: AuthSources{
								Header: &AuthSource{
									Enabled: true,
								},
							},
						},
					},
				},
			},
			Upstream: Upstream{
				URL: "https://example.org/api",
			},
			Info: Info{
				Name: "New OAS API",
				State: State{
					Active: true,
				},
			},
		}

		assert.Equal(t, expectedTykExtension, *oasDef.GetTykExtension())
	})

	t.Run("error when supplied invalid upstreamURL param", func(t *testing.T) {
		oasDef := OAS{
			T: openapi3.T{
				Info: &openapi3.Info{
					Title: "OAS API",
				},
				Servers: openapi3.Servers{
					{
						URL: "https://example-org.com/api",
					},
				},
			},
		}

		existingTykExtension := XTykAPIGateway{
			Info: Info{
				Name: "New OAS API",
			},
			Server: Server{
				ListenPath: ListenPath{
					Value: "/listen-api",
				},
			},
		}

		oasDef.SetTykExtension(&existingTykExtension)

		err := oasDef.BuildDefaultTykExtension(TykExtensionConfigParams{
			ListenPath:  "/new-listen-api",
			UpstreamURL: "invalid-url",
		})
		assert.ErrorIs(t, err, errInvalidUpstreamURL)
	})

	t.Run("error when no supplied params and invalid URL in servers", func(t *testing.T) {
		oasDef := OAS{
			T: openapi3.T{
				Info: &openapi3.Info{
					Title: "OAS API",
				},
				Servers: openapi3.Servers{
					{
						URL: "/listen-api",
					},
				},
			},
		}

		existingTykExtension := XTykAPIGateway{
			Info: Info{
				Name: "New OAS API",
			},
			Server: Server{
				ListenPath: ListenPath{
					Value: "/listen-api",
				},
			},
		}

		oasDef.SetTykExtension(&existingTykExtension)

		err := oasDef.BuildDefaultTykExtension(TykExtensionConfigParams{})
		assert.ErrorIs(t, err, errInvalidServerURL)
	})

	t.Run("error when no supplied params and no servers", func(t *testing.T) {
		oasDef := OAS{
			T: openapi3.T{
				Info: &openapi3.Info{
					Title: "OAS API",
				},
			},
		}

		existingTykExtension := XTykAPIGateway{
			Info: Info{
				Name: "New OAS API",
			},
			Server: Server{
				ListenPath: ListenPath{
					Value: "/listen-api",
				},
			},
		}

		oasDef.SetTykExtension(&existingTykExtension)

		err := oasDef.BuildDefaultTykExtension(TykExtensionConfigParams{})
		assert.ErrorIs(t, err, errEmptyServersObject)
	})

	t.Run("middlewares", func(t *testing.T) {
		trueVal, falseVal := true, false

		const (
			tykGetOperationID  = "petsGET"
			tykPostOperationID = "petsPOST"
			oasGetOperationID  = "getPets"
			oasPostOperationID = "postPets"
		)
		getOASDef := func(withOperationID bool) OAS {
			oasDef := OAS{
				T: openapi3.T{
					Info: &openapi3.Info{
						Title: "OAS API",
					},
					Servers: openapi3.Servers{
						{
							URL: "https://example-org.com/api",
						},
					},
					Paths: openapi3.Paths{
						"/pets": {
							Get: &openapi3.Operation{
								Responses: openapi3.Responses{},
							},
							Post: &openapi3.Operation{
								Responses: openapi3.Responses{},
							},
						},
					},
				},
			}

			if withOperationID {
				oasDef.Paths = openapi3.Paths{
					"/pets": {
						Get: &openapi3.Operation{
							OperationID: "getPets",
							Responses:   openapi3.Responses{},
						},
						Post: &openapi3.Operation{
							OperationID: oasPostOperationID,
							Responses:   openapi3.Responses{},
						},
					},
				}
			}

			return oasDef
		}

		fillReqBody := func(oasDef *OAS, path, method string) {
			pathItem := oasDef.Paths.Find(path)
			oasOperation := pathItem.GetOperation(method)
			reqBody := openapi3.NewRequestBody()
			reqBody.Description = "JSON req body"
			valueSchema := openapi3.NewSchema()
			valueSchema.Properties = openapi3.Schemas{
				"value": {
					Value: &openapi3.Schema{
						Type: openapi3.TypeBoolean,
					},
				},
			}
			content := openapi3.NewContentWithSchema(valueSchema, []string{contentTypeJSON})
			reqBody.Content = content
			oasOperation.RequestBody = &openapi3.RequestBodyRef{Value: reqBody}
		}

		getExpectedOperations := func(enabled bool, middleware string, oasOperationID bool) Operations {
			if middleware == MiddlewareAllowList && oasOperationID {
				return Operations{
					oasGetOperationID: {
						Allow: &Allowance{
							Enabled: enabled,
						},
					},
					oasPostOperationID: {
						Allow: &Allowance{
							Enabled: enabled,
						},
					},
				}
			} else if middleware == MiddlewareAllowList && !oasOperationID {
				return Operations{
					tykGetOperationID: {
						Allow: &Allowance{
							Enabled: enabled,
						},
					},
					tykPostOperationID: {
						Allow: &Allowance{
							Enabled: enabled,
						},
					},
				}
			}

			if middleware == MiddlewareValidateRequest && oasOperationID {
				return Operations{
					oasPostOperationID: {
						ValidateRequest: &ValidateRequest{
							Enabled:           enabled,
							ErrorResponseCode: http.StatusBadRequest,
						},
					},
				}
			} else if middleware == MiddlewareValidateRequest && !oasOperationID {
				return Operations{
					tykPostOperationID: {
						ValidateRequest: &ValidateRequest{
							Enabled:           enabled,
							ErrorResponseCode: http.StatusBadRequest,
						},
					},
				}
			}

			return nil
		}

		t.Run("allowList", func(t *testing.T) {
			t.Run("enable allowList for all paths when no configured operationID in OAS", func(t *testing.T) {
				oasDef := getOASDef(false)
				expectedOperations := getExpectedOperations(true, MiddlewareAllowList, false)

				tykExtensionConfigParams := TykExtensionConfigParams{
					AllowList: &trueVal,
				}

				err := oasDef.BuildDefaultTykExtension(tykExtensionConfigParams)

				assert.NoError(t, err)
				assert.Equal(t, expectedOperations, oasDef.GetTykExtension().Middleware.Operations)
			})

			t.Run("enable allowList for all paths when operationID is configured in OAS", func(t *testing.T) {
				oasDef := getOASDef(true)

				expectedOperations := getExpectedOperations(true, MiddlewareAllowList, true)

				tykExtensionConfigParams := TykExtensionConfigParams{
					AllowList: &trueVal,
				}

				err := oasDef.BuildDefaultTykExtension(tykExtensionConfigParams)

				assert.NoError(t, err)
				assert.Equal(t, expectedOperations, oasDef.GetTykExtension().Middleware.Operations)
			})

			t.Run("disable allowList for all paths when no configured operationID in OAS", func(t *testing.T) {
				oasDef := getOASDef(false)

				expectedOperations := getExpectedOperations(false, MiddlewareAllowList, false)

				tykExtensionConfigParams := TykExtensionConfigParams{
					AllowList: &falseVal,
				}

				err := oasDef.BuildDefaultTykExtension(tykExtensionConfigParams)

				assert.NoError(t, err)
				assert.Equal(t, expectedOperations, oasDef.GetTykExtension().Middleware.Operations)
			})

			t.Run("disable allowList for all paths when operationID is configured in OAS", func(t *testing.T) {
				oasDef := getOASDef(true)

				expectedOperations := getExpectedOperations(false, MiddlewareAllowList, true)

				tykExtensionConfigParams := TykExtensionConfigParams{
					AllowList: &falseVal,
				}

				err := oasDef.BuildDefaultTykExtension(tykExtensionConfigParams)

				assert.NoError(t, err)
				assert.Equal(t, expectedOperations, oasDef.GetTykExtension().Middleware.Operations)
			})

			t.Run("override allowList (disable) configured in tyk extension - do not toggle block list if any", func(t *testing.T) {
				oasDef := getOASDef(true)

				tykExt := XTykAPIGateway{
					Server: Server{
						ListenPath: ListenPath{
							Value: "/",
						},
					},
					Upstream: Upstream{
						URL: "https://example-org.com/api",
					},
					Info: Info{
						Name: "OAS API",
						State: State{
							Active: true,
						},
					},
					Middleware: &Middleware{
						Operations: Operations{
							oasGetOperationID: {
								Allow: &Allowance{
									Enabled: true,
								},
								Block: &Allowance{
									Enabled: false,
								},
							},
							oasPostOperationID: {
								Allow: &Allowance{
									Enabled: true,
								},
								Block: &Allowance{
									Enabled: false,
								},
							},
						},
					},
				}

				expectedOperations := Operations{
					oasGetOperationID: {
						Allow: &Allowance{
							Enabled: false,
						},
						Block: &Allowance{
							Enabled: false,
						},
					},
					oasPostOperationID: {
						Allow: &Allowance{
							Enabled: false,
						},
						Block: &Allowance{
							Enabled: false,
						},
					},
				}

				oasDef.SetTykExtension(&tykExt)

				tykExtensionConfigParams := TykExtensionConfigParams{
					AllowList: &falseVal,
				}

				err := oasDef.BuildDefaultTykExtension(tykExtensionConfigParams)

				assert.NoError(t, err)

				assert.Equal(t, expectedOperations, oasDef.GetTykExtension().Middleware.Operations)
			})

			t.Run("override allowList (enable) configured in tyk extension - toggle enabled block list if any", func(t *testing.T) {
				oasDef := getOASDef(true)

				tykExt := XTykAPIGateway{
					Server: Server{
						ListenPath: ListenPath{
							Value: "/",
						},
					},
					Upstream: Upstream{
						URL: "https://example-org.com/api",
					},
					Info: Info{
						Name: "OAS API",
						State: State{
							Active: true,
						},
					},
					Middleware: &Middleware{
						Operations: Operations{
							oasGetOperationID: {
								Allow: &Allowance{
									Enabled: false,
								},
								Block: &Allowance{
									Enabled: true,
								},
							},
							oasPostOperationID: {
								Allow: &Allowance{
									Enabled: false,
								},
							},
						},
					},
				}

				expectedOperations := Operations{
					oasGetOperationID: {
						Allow: &Allowance{
							Enabled: true,
						},
						Block: &Allowance{
							Enabled: false,
						},
					},
					oasPostOperationID: {
						Allow: &Allowance{
							Enabled: true,
						},
					},
				}

				oasDef.SetTykExtension(&tykExt)

				tykExtensionConfigParams := TykExtensionConfigParams{
					AllowList: &trueVal,
				}

				err := oasDef.BuildDefaultTykExtension(tykExtensionConfigParams)

				assert.NoError(t, err)

				assert.Equal(t, expectedOperations, oasDef.GetTykExtension().Middleware.Operations)
			})

			t.Run("do not configure allowList when parameter is not provided (nil)", func(t *testing.T) {
				oasDef := getOASDef(false)

				tykExt := XTykAPIGateway{
					Server: Server{
						ListenPath: ListenPath{
							Value: "/",
						},
					},
					Upstream: Upstream{
						URL: "https://example-org.com/api",
					},
					Info: Info{
						Name: "OAS API",
						State: State{
							Active: true,
						},
					},
					Middleware: &Middleware{
						Operations: Operations{
							tykGetOperationID: {
								Allow: &Allowance{
									Enabled: false,
								},
							},
							tykPostOperationID: {
								Allow: &Allowance{
									Enabled: true,
								},
							},
						},
					},
				}

				oasDef.SetTykExtension(&tykExt)

				var expectedOperations = make(Operations)
				for k, v := range tykExt.Middleware.Operations {
					expectedOperations[k] = v
				}

				tykExtensionConfigParams := TykExtensionConfigParams{
					AllowList: nil,
				}

				err := oasDef.BuildDefaultTykExtension(tykExtensionConfigParams)

				assert.NoError(t, err)
				assert.EqualValues(t, expectedOperations, oasDef.GetTykExtension().Middleware.Operations)
			})
		})

		t.Run("validateRequest", func(t *testing.T) {

			t.Run("do not configure validateRequest for paths where request body is not specified for application/json",
				func(t *testing.T) {
					oasDef := getOASDef(false)
					fillReqBody(&oasDef, "/pets", http.MethodPost)

					expectedOperations := getExpectedOperations(true, MiddlewareValidateRequest, false)

					tykExtensionConfigParams := TykExtensionConfigParams{
						ValidateRequest: &trueVal,
					}

					err := oasDef.BuildDefaultTykExtension(tykExtensionConfigParams)

					assert.NoError(t, err)
					assert.EqualValues(t, expectedOperations, oasDef.GetTykExtension().Middleware.Operations)
				})

			t.Run("enable validateRequest for all paths with application/json req body when no configured operationID in OAS",
				func(t *testing.T) {
					oasDef := getOASDef(false)
					fillReqBody(&oasDef, "/pets", http.MethodPost)

					expectedOperations := getExpectedOperations(true, MiddlewareValidateRequest, false)

					tykExtensionConfigParams := TykExtensionConfigParams{
						ValidateRequest: &trueVal,
					}

					err := oasDef.BuildDefaultTykExtension(tykExtensionConfigParams)

					assert.NoError(t, err)
					assert.Equal(t, expectedOperations, oasDef.GetTykExtension().Middleware.Operations)
				})

			t.Run("enable validateRequest for all paths with application/json req body when operationID is configured in OAS",
				func(t *testing.T) {
					oasDef := getOASDef(true)
					fillReqBody(&oasDef, "/pets", http.MethodPost)

					expectedOperations := getExpectedOperations(true, MiddlewareValidateRequest, true)

					tykExtensionConfigParams := TykExtensionConfigParams{
						ValidateRequest: &trueVal,
					}

					err := oasDef.BuildDefaultTykExtension(tykExtensionConfigParams)

					assert.NoError(t, err)
					assert.Equal(t, expectedOperations, oasDef.GetTykExtension().Middleware.Operations)
				})

			t.Run("disable validateRequest for all paths with application/json req body when no configured operationID in OAS",
				func(t *testing.T) {
					oasDef := getOASDef(false)
					fillReqBody(&oasDef, "/pets", http.MethodPost)
					expectedOperations := getExpectedOperations(false, MiddlewareValidateRequest, false)

					tykExtensionConfigParams := TykExtensionConfigParams{
						ValidateRequest: &falseVal,
					}

					err := oasDef.BuildDefaultTykExtension(tykExtensionConfigParams)

					assert.NoError(t, err)
					assert.Equal(t, expectedOperations, oasDef.GetTykExtension().Middleware.Operations)
				})

			t.Run("disable validateRequest for all paths with application/json req body when operationID is configured in OAS",
				func(t *testing.T) {
					oasDef := getOASDef(true)
					fillReqBody(&oasDef, "/pets", http.MethodPost)

					expectedOperations := getExpectedOperations(false, MiddlewareValidateRequest, true)

					tykExtensionConfigParams := TykExtensionConfigParams{
						ValidateRequest: &falseVal,
					}

					err := oasDef.BuildDefaultTykExtension(tykExtensionConfigParams)

					assert.NoError(t, err)
					assert.Equal(t, expectedOperations, oasDef.GetTykExtension().Middleware.Operations)
				})

			t.Run("override validateRequest configured in tyk extension", func(t *testing.T) {
				oasDef := getOASDef(true)
				fillReqBody(&oasDef, "/pets", http.MethodPost)

				tykExt := XTykAPIGateway{
					Server: Server{
						ListenPath: ListenPath{
							Value: "/",
						},
					},
					Upstream: Upstream{
						URL: "https://example-org.com/api",
					},
					Info: Info{
						Name: "OAS API",
						State: State{
							Active: true,
						},
					},
					Middleware: &Middleware{
						Operations: Operations{
							oasPostOperationID: {
								ValidateRequest: &ValidateRequest{
									Enabled: false,
								},
							},
						},
					},
				}

				oasDef.SetTykExtension(&tykExt)

				expectedOperations := getExpectedOperations(true, MiddlewareValidateRequest, true)

				tykExtensionConfigParams := TykExtensionConfigParams{
					ValidateRequest: &trueVal,
				}

				err := oasDef.BuildDefaultTykExtension(tykExtensionConfigParams)

				assert.NoError(t, err)

				assert.Equal(t, expectedOperations, oasDef.GetTykExtension().Middleware.Operations)
			})

			t.Run("do not configure validateRequest when parameter is not provided (nil)", func(t *testing.T) {
				oasDef := getOASDef(true)

				tykExt := XTykAPIGateway{
					Server: Server{
						ListenPath: ListenPath{
							Value: "/",
						},
					},
					Upstream: Upstream{
						URL: "https://example-org.com/api",
					},
					Info: Info{
						Name: "OAS API",
						State: State{
							Active: true,
						},
					},
					Middleware: &Middleware{
						Operations: Operations{
							oasGetOperationID: {
								ValidateRequest: &ValidateRequest{
									Enabled: true,
								},
								Allow: &Allowance{
									Enabled: false,
								},
							},
							oasPostOperationID: {
								ValidateRequest: &ValidateRequest{
									Enabled: false,
								},
								Allow: &Allowance{
									Enabled: true,
								},
							},
						},
					},
				}

				oasDef.SetTykExtension(&tykExt)

				var expectedOperations = make(Operations)
				for k, v := range tykExt.Middleware.Operations {
					expectedOperations[k] = v
				}

				tykExtensionConfigParams := TykExtensionConfigParams{
					ValidateRequest: nil,
				}

				err := oasDef.BuildDefaultTykExtension(tykExtensionConfigParams)

				assert.NoError(t, err)
				actualTykExtension := oasDef.GetTykExtension()
				assert.EqualValues(t, expectedOperations, actualTykExtension.Middleware.Operations)
			})

			t.Run("do not configure validateRequest when no paths have application/json req body",
				func(t *testing.T) {
					oasDef := getOASDef(true)

					tykExtensionConfigParams := TykExtensionConfigParams{
						ValidateRequest: &falseVal,
					}

					err := oasDef.BuildDefaultTykExtension(tykExtensionConfigParams)

					assert.NoError(t, err)
					assert.Nil(t, oasDef.GetTykExtension().Middleware)
				})
		})

	})

}

func TestGetTykExtensionConfigParams(t *testing.T) {
	trueVal, falseVal := true, false
	t.Run("extract all params when provided", func(t *testing.T) {
		endpoint, err := url.Parse("/")
		assert.NoError(t, err)

		listenPath := "/listen-api"
		upstreamURL := "https://upstream.org"
		customDomain := "custom-domain.org"

		queryParams := endpoint.Query()
		queryParams.Set("listenPath", listenPath)
		queryParams.Set("upstreamURL", upstreamURL)
		queryParams.Set("customDomain", customDomain)
		queryParams.Set("validateRequest", "true")
		queryParams.Set("allowList", "false")

		endpoint.RawQuery = queryParams.Encode()
		r, err := http.NewRequest(http.MethodPatch, endpoint.String(), nil)
		assert.NoError(t, err)

		tykExtConfigParams := GetTykExtensionConfigParams(r)

		expectedConfigParams := TykExtensionConfigParams{
			ListenPath:      listenPath,
			UpstreamURL:     upstreamURL,
			CustomDomain:    customDomain,
			AllowList:       &falseVal,
			ValidateRequest: &trueVal,
		}

		assert.Equal(t, &expectedConfigParams, tykExtConfigParams)
	})

	t.Run("nil when no params provided", func(t *testing.T) {
		endpoint, err := url.Parse("/")
		assert.NoError(t, err)

		r, err := http.NewRequest(http.MethodPatch, endpoint.String(), nil)
		assert.NoError(t, err)

		assert.Nil(t, GetTykExtensionConfigParams(r))
	})

	t.Run("nil for middleware when params not provided", func(t *testing.T) {
		endpoint, err := url.Parse("/")
		assert.NoError(t, err)

		listenPath := "/listen-api"
		upstreamURL := "https://upstream.org"
		customDomain := "custom-domain.org"

		queryParams := endpoint.Query()
		queryParams.Set("listenPath", listenPath)
		queryParams.Set("upstreamURL", upstreamURL)
		queryParams.Set("customDomain", customDomain)

		endpoint.RawQuery = queryParams.Encode()
		r, err := http.NewRequest(http.MethodPatch, endpoint.String(), nil)
		assert.NoError(t, err)

		tykExtConfigParams := GetTykExtensionConfigParams(r)

		expectedConfigParams := TykExtensionConfigParams{
			ListenPath:   listenPath,
			UpstreamURL:  upstreamURL,
			CustomDomain: customDomain,
		}

		assert.Equal(t, &expectedConfigParams, tykExtConfigParams)
	})

	t.Run("not nil when at least one parameter is provided", func(t *testing.T) {
		endpoint, err := url.Parse("/")
		assert.NoError(t, err)

		queryParams := endpoint.Query()
		queryParams.Set("allowList", "true")

		endpoint.RawQuery = queryParams.Encode()
		r, err := http.NewRequest(http.MethodPatch, endpoint.String(), nil)
		assert.NoError(t, err)

		tykExtConfigParams := GetTykExtensionConfigParams(r)

		expectedConfigParams := TykExtensionConfigParams{
			AllowList: &trueVal,
		}

		assert.Equal(t, &expectedConfigParams, tykExtConfigParams)
	})
}

func TestOAS_importAuthentication(t *testing.T) {
	const (
		testSecurityNameToken = "my_auth_token"
		testHeaderName        = "my-auth-token-header"
		testCookieName        = "my-auth-token-cookie"
	)

	t.Run("security is empty", func(t *testing.T) {
		oas := OAS{}
		oas.SetTykExtension(&XTykAPIGateway{})

		err := oas.importAuthentication(true)
		assert.ErrorIs(t, errEmptySecurityObject, err)

		authentication := oas.getTykAuthentication()
		assert.Nil(t, authentication)
	})

	t.Run("add authentication", func(t *testing.T) {
		check := func(t *testing.T, enable bool) {
			oas := OAS{}
			oas.Security = openapi3.SecurityRequirements{
				{testSecurityNameToken: []string{}},
			}

			tokenScheme := openapi3.NewSecurityScheme()
			tokenScheme.Type = typeApiKey
			tokenScheme.In = cookie
			tokenScheme.Name = testCookieName

			jwtScheme := openapi3.NewSecurityScheme()
			jwtScheme.Type = typeHttp
			jwtScheme.Scheme = schemeBearer
			jwtScheme.BearerFormat = bearerFormatJWT

			oas.Components.SecuritySchemes = openapi3.SecuritySchemes{
				testSecurityNameToken: &openapi3.SecuritySchemeRef{
					Value: tokenScheme,
				},
			}

			oas.SetTykExtension(&XTykAPIGateway{})

			err := oas.importAuthentication(enable)
			assert.NoError(t, err)

			authentication := oas.getTykAuthentication()

			assert.Equal(t, authentication.Enabled, enable)

			expectedSecuritySchemes := SecuritySchemes{
				testSecurityNameToken: &Token{
					Enabled: enable,
					AuthSources: AuthSources{
						Cookie: &AuthSource{
							Enabled: true,
						},
					},
				},
			}

			assert.Equal(t, expectedSecuritySchemes, authentication.SecuritySchemes)
		}

		t.Run("enable=true", func(t *testing.T) {
			check(t, true)
		})

		t.Run("enable=false", func(t *testing.T) {
			check(t, false)
		})

	})

	t.Run("update existing one", func(t *testing.T) {
		oas := OAS{}
		oas.Security = openapi3.SecurityRequirements{
			{testSecurityNameToken: []string{}},
		}

		securityScheme := openapi3.NewSecurityScheme()
		securityScheme.Type = typeApiKey
		securityScheme.In = cookie
		securityScheme.Name = testCookieName

		oas.Components.SecuritySchemes = openapi3.SecuritySchemes{
			testSecurityNameToken: &openapi3.SecuritySchemeRef{
				Value: securityScheme,
			},
		}

		xTykAPIGateway := &XTykAPIGateway{
			Server: Server{
				Authentication: &Authentication{
					SecuritySchemes: SecuritySchemes{
						testSecurityNameToken: &Token{
							Enabled: false,
							AuthSources: AuthSources{
								Header: &AuthSource{
									Enabled: true,
									Name:    testHeaderName,
								},
							},
						},
					},
				},
			},
		}

		oas.SetTykExtension(xTykAPIGateway)

		err := oas.importAuthentication(true)
		assert.NoError(t, err)

		authentication := oas.getTykAuthentication()

		assert.True(t, authentication.Enabled)

		expectedSecuritySchemes := SecuritySchemes{
			testSecurityNameToken: &Token{
				Enabled: true,
				AuthSources: AuthSources{
					Header: &AuthSource{
						Enabled: true,
						Name:    testHeaderName,
					},
					Cookie: &AuthSource{
						Enabled: true,
					},
				},
			},
		}

		assert.Equal(t, expectedSecuritySchemes, authentication.SecuritySchemes)
	})
}

func TestSecuritySchemes_Import(t *testing.T) {
	const (
		testSecurityNameToken = "my_auth_token"
		testHeaderName        = "my-auth-token-header"
		testCookieName        = "my-auth-token-cookie"
	)

	t.Run("token", func(t *testing.T) {
		check := func(t *testing.T, enable bool) {
			securitySchemes := SecuritySchemes{}
			nativeSecurityScheme := &openapi3.SecurityScheme{
				Type: typeApiKey,
				In:   header,
				Name: testHeaderName,
			}

			err := securitySchemes.Import(testSecurityNameToken, nativeSecurityScheme, enable)
			assert.NoError(t, err)

			expectedToken := &Token{
				Enabled: enable,
				AuthSources: AuthSources{
					Header: &AuthSource{
						Enabled: true,
					},
				},
			}

			assert.Equal(t, expectedToken, securitySchemes[testSecurityNameToken])
		}

		t.Run("enable=true", func(t *testing.T) {
			check(t, true)
		})

		t.Run("enable=false", func(t *testing.T) {
			check(t, false)
		})
	})

	t.Run("update existing one", func(t *testing.T) {
		existingToken := &Token{
			AuthSources: AuthSources{
				Cookie: &AuthSource{
					Enabled: true,
					Name:    testCookieName,
				},
			},
		}
		securitySchemes := SecuritySchemes{
			testSecurityNameToken: existingToken,
		}

		nativeSecurityScheme := &openapi3.SecurityScheme{
			Type: typeApiKey,
			In:   header,
			Name: testHeaderName,
		}

		err := securitySchemes.Import(testSecurityNameToken, nativeSecurityScheme, true)
		assert.NoError(t, err)

		expectedToken := &Token{
			Enabled: true,
			AuthSources: AuthSources{
				Header: &AuthSource{
					Enabled: true,
				},
				Cookie: &AuthSource{
					Enabled: true,
					Name:    testCookieName,
				},
			},
		}

		assert.Equal(t, expectedToken, securitySchemes[testSecurityNameToken])
	})
}

func TestAuthSources_Import(t *testing.T) {
	expectedAuthSource := &AuthSource{Enabled: true}

	t.Run(header, func(t *testing.T) {
		as := AuthSources{}
		as.Import(header)

		assert.Equal(t, expectedAuthSource, as.Header)
	})

	t.Run(query, func(t *testing.T) {
		as := AuthSources{}
		as.Import(query)

		assert.Equal(t, expectedAuthSource, as.Query)
	})

	t.Run(cookie, func(t *testing.T) {
		as := AuthSources{}
		as.Import(cookie)

		assert.Equal(t, expectedAuthSource, as.Cookie)
	})
}