window.ui = SwaggerUIBundle({
  url: './openapi.yaml',
  dom_id: '#swagger-ui',
  deepLinking: true,
  persistAuthorization: true,
  presets: [SwaggerUIBundle.presets.apis, SwaggerUIStandalonePreset],
  layout: 'StandaloneLayout',
})
