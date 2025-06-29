"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const jsdom_1 = require("jsdom");
const fs_1 = require("fs");
const assert_1 = require("assert");
const webcrypto_1 = require("@peculiar/webcrypto");
if (typeof globalThis.crypto === 'undefined') {
    globalThis.crypto = new webcrypto_1.Crypto();
}
require('./lib/js/wasm_exec.js');
class CheckResults {
    constructor() {
        this.errors = null;
        this.resolve = null;
    }
    onCheckCompleted(errs) {
        this.errors = errs;
        if (this.resolve !== null) {
            this.resolve(errs);
            this.resolve = null;
        }
    }
    waitCheckCompleted() {
        return new Promise(resolve => {
            if (this.errors !== null) {
                resolve(this.errors);
                return;
            }
            this.resolve = resolve;
        });
    }
    reset() {
        this.errors = null;
    }
}
describe('main.wasm', function () {
    const results = new CheckResults();
    before(async function () {
        const dom = new jsdom_1.JSDOM('');
        dom.window.dismissLoading = function () {
        };
        dom.window.getYamlSource = function () {
            return `
on: push

jobs:
  test:
    steps:
      - run: echo 'hi'`;
        };
        dom.window.onCheckCompleted = results.onCheckCompleted.bind(results);
        global.window = dom.window;
        const go = new Go();
        const bin = await fs_1.promises.readFile('./main.wasm');
        const result = await WebAssembly.instantiate(bin, go.importObject);
        go.run(result.instance);
    });
    it('shows first result on loading', async function () {
        const errors = await results.waitCheckCompleted();
        const json = JSON.stringify(errors);
        assert_1.strict.equal(errors.length, 1, json);
        const err = errors[0];
        assert_1.strict.equal(err.message, '"runs-on" section is missing in job "test"', `message is unexpected: ${json}`);
        assert_1.strict.equal(err.line, 5, `line is unexpected: ${json}`);
        assert_1.strict.equal(err.column, 3, `column is unexpected: ${json}`);
        assert_1.strict.equal(err.kind, 'syntax-check', `kind is unexpected: ${json}`);
    });
    it('reports some errors by running actionlint with runActionlint', async function () {
        assert_1.strict.ok(window.runActionlint);
        results.reset();
        const source = `
on: foo

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - run: echo 'hi'`;
        window.runActionlint(source);
        const errors = await results.waitCheckCompleted();
        const json = JSON.stringify(errors);
        assert_1.strict.equal(errors.length, 1, json);
        const err = errors[0];
        assert_1.strict.ok(err.message.includes('unknown Webhook event "foo"'), `message is unexpected: ${json}`);
        assert_1.strict.equal(err.line, 2, `line is unexpected: ${json}`);
        assert_1.strict.equal(err.column, 5, `column is unexpected: ${json}`);
        assert_1.strict.equal(err.kind, 'events', `kind is unexpected: ${json}`);
    });
    it('reports no error by running actionlint with runActionlint', async function () {
        assert_1.strict.ok(window.runActionlint);
        results.reset();
        const source = `
on: push

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - run: echo 'hi'`;
        window.runActionlint(source);
        const errors = await results.waitCheckCompleted();
        const json = JSON.stringify(errors);
        assert_1.strict.equal(errors.length, 0, json);
    });
});
//# sourceMappingURL=test.js.map