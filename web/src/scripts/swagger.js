import {fileURLToPath} from 'node:url';
import {generateApi} from 'swagger-typescript-api';

await generateApi({
    input: fileURLToPath(new URL('../../../docs/swagger.json', import.meta.url)),
    output: fileURLToPath(new URL('../api', import.meta.url)),
    fileName: 'api.ts',
    generateClient: true,
    httpClientType: 'axios',
    moduleNameFirstTag: true,
    extractRequestParams: true,
    extractRequestBody: true,
    extractResponseBody: true,
    hooks: {
        onCreateRoute(route) {
            // Swagger 2 formData is otherwise treated as multipart by the generator.
            if (route.requestBodyInfo.contentTypes.includes('application/x-www-form-urlencoded')) {
                route.requestBodyInfo.contentKind = 'URL_ENCODED';
                route.request.formData = false;
                route.request.isQueryBody = true;
            }

            return route;
        },
    },
});
