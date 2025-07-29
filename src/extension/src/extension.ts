import * as vscode from 'vscode';
import * as lspClient from './lspclient';

export async function activate(context: vscode.ExtensionContext) {
	if (await lspClient.getUlspEnablementStatus()) {
      const client = new lspClient.LSPClient()
      await client.activate(context)
    }
}

export function deactivate() {}
