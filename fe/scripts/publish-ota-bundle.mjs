import { createHash } from 'node:crypto';
import { existsSync, mkdirSync, readFileSync, writeFileSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { execFileSync } from 'node:child_process';
import { fileURLToPath } from 'node:url';

const scriptDir = dirname(fileURLToPath(import.meta.url));
const projectDir = resolve(scriptDir, '..');
const workspaceDir = resolve(projectDir, '..');
const buildDir = resolve(projectDir, 'build');
const channel = process.env.OTA_CHANNEL || 'dev';
const nativeVersion = process.env.NATIVE_VERSION || '1.0';
const version =
	process.env.OTA_VERSION ||
	new Date()
		.toISOString()
		.replace(/[-:]/g, '')
		.replace(/\.\d{3}Z$/, 'Z');
const outputDir = resolve(workspaceDir, 'be', 'storage', 'ota', 'android', channel);
const zipName = `${version}.zip`;
const zipPath = resolve(outputDir, zipName);
const metadataPath = resolve(outputDir, 'latest.json');

if (!existsSync(buildDir)) {
	throw new Error('Missing build directory. Run `yarn build` before publishing OTA.');
}

mkdirSync(outputDir, { recursive: true });
execFileSync('zip', ['-qr', zipPath, '.'], { cwd: buildDir, stdio: 'inherit' });

const checksum = createHash('sha256').update(readFileSync(zipPath)).digest('hex');
const metadata = {
	version,
	url: `/ota/android/${channel}/${zipName}`,
	checksum,
	mandatory: false,
	min_native_version: nativeVersion,
	max_native_version: nativeVersion,
	notes: process.env.OTA_NOTES || ''
};

writeFileSync(metadataPath, `${JSON.stringify(metadata, null, 2)}\n`);

console.log(`Published Android OTA bundle ${version}`);
console.log(`Bundle: ${zipPath}`);
console.log(`Metadata: ${metadataPath}`);
