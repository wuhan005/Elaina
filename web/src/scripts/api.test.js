import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import {createRequire} from 'node:module';
import test from 'node:test';
import ts from 'typescript';

const source = readFileSync(new URL('../api/api.ts', import.meta.url), 'utf8');
const compiled = ts.transpileModule(source, {
    compilerOptions: {
        module: ts.ModuleKind.CommonJS,
        target: ts.ScriptTarget.ES2022,
        esModuleInterop: true,
    },
}).outputText;
const api = {};
new Function('require', 'exports', compiled)(createRequire(import.meta.url), api);

test('the generated execution client sends URL-encoded source code', async () => {
    const client = new api.Api({
        adapter: async (config) => ({
            data: {},
            status: 200,
            statusText: 'OK',
            headers: {},
            config,
        }),
    });
    const code = 'print("+&= 中文")\n';
    const response = await client.runner.executeCode({uid: 'sandbox'}, {
        lang: 'javascript',
        code,
    });

    assert.equal(response.config.url, '/r/sandbox/execute');
    assert.equal(response.config.headers.get('Content-Type'), 'application/x-www-form-urlencoded');
    const body = new URLSearchParams(response.config.data);
    assert.equal(body.get('lang'), 'javascript');
    assert.equal(body.get('code'), code);
});
