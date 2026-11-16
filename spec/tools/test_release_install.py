"""Negative controls for real-install acceptance; no fabricated release passes."""
import copy
import json
import os
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch
import release_install as gate

class InstallGateTests(unittest.TestCase):
    def test_only_explicit_policy_versions(self):
        for v in ('v0.1.0-rc.1', 'v0.2.0', 'v0.1.12'):
            self.assertEqual(gate.version(v), v)
        for v in ('latest', 'main', 'v1.0.0', 'v0.0.0-diameterkit.batch85', 'v0.1.0+local', 'v0.01.0', 'v0.1.0-rc.0', 'v0.1.0-rc.01', 'v0.1.0\n'):
            with self.subTest(v=v), self.assertRaises(ValueError): gate.version(v)

    def test_host_workspace_bypasses_and_warm_caches_not_inherited(self):
        with tempfile.TemporaryDirectory() as path, patch.dict(os.environ, {
            'GOWORK':'/old/go.work','GOMODCACHE':'/old/mod','GOCACHE':'/old/build',
            'GOPROXY':'file:///local','GOSUMDB':'off','GONOSUMDB':'*','GOPRIVATE':'*',
            'GOFLAGS':'-mod=vendor','GOOS':'windows','GOARCH':'386','GOENV':'/custom'}):
            env = gate.environment(Path(path))
            for key,value in {'GOWORK':'off','GOSUMDB':'sum.golang.org','GOPROXY':'https://proxy.golang.org','GONOSUMDB':'','GOPRIVATE':'','GOFLAGS':'-mod=readonly'}.items():
                self.assertEqual(env[key],value)
            self.assertNotIn('GOARCH',env)
            self.assertNotIn('GOOS',env)
            for name in ('GOMODCACHE','GOCACHE','GOPATH'):
                self.assertTrue(Path(env[name]).is_relative_to(path))
                self.assertEqual(list(Path(env[name]).iterdir()),[])

    def test_manifests_reject_replace_exclude_and_core_drift(self):
        gate.validate_manifest({'Module':{'Path':gate.CORE},'Require':None},gate.CORE,{})
        expected = {gate.CORE:'v0.1.0-rc.1',gate.UPSTREAM:'v0.8.0'}
        valid = {'Module':{'Path':gate.ADAPTER},'Require':[{'Path':k,'Version':v} for k,v in expected.items()]}
        gate.validate_manifest(valid,gate.ADAPTER,expected)
        for change in ({'Replace':[{'Old':{'Path':gate.CORE},'New':{'Path':'/checkout'}}]}, {'Exclude':[{'Path':gate.CORE}]}, {'Require':[]}):
            with self.assertRaises(RuntimeError):gate.validate_manifest({**valid,**change},gate.ADAPTER,expected)

    def test_graph_rejects_mvs_upgrade_replace_and_checkout(self):
        with tempfile.TemporaryDirectory() as path:
            root=Path(path);expected={gate.CORE:'v0.1.0-rc.1'}
            rows=[{'Path':gate.CONSUMER,'Main':True,'Dir':str(root/'consumer')},{'Path':gate.CORE,'Version':'v0.1.0-rc.1','Dir':str(root/'mod/core')}]
            def check(items):return gate.validate_graph('\n'.join(json.dumps(r) for r in items),expected,root/'mod',root/'consumer')
            check(rows)
            for change in ({'Main':True},{'Replace':{'Dir':'/checkout'}},{'Version':'v0.2.0'},{'Dir':str(root/'checkout')}):
                bad=copy.deepcopy(rows);bad[1].update(change)
                with self.subTest(change=change),self.assertRaises(RuntimeError):check(bad)
            with self.assertRaises(RuntimeError):check(rows+[{'Path':'unexpected.example/module'}])

    def test_download_requires_checksums_and_private_paths(self):
        with tempfile.TemporaryDirectory() as path:
            cache=Path(path)
            row={'Path':gate.CORE,'Version':'v0.1.0-rc.1','Sum':'h1:'+'a'*43+'=','GoModSum':'h1:'+'b'*43+'='}
            row.update({key:str(cache/key) for key in ('Dir','GoMod','Info','Zip')})
            gate.validate_download(row,gate.CORE,'v0.1.0-rc.1',cache)
            for change in ({'Sum':''},{'GoModSum':''},{'Dir':'/checkout'},{'Error':'404'},{'Version':'v0.2.0'}):
                with self.subTest(change=change),self.assertRaises(RuntimeError):gate.validate_download({**row,**change},gate.CORE,'v0.1.0-rc.1',cache)

    def test_missing_release_is_failed_with_logs(self):
        with tempfile.TemporaryDirectory() as path:
            root=Path(path);output=root/'logs';output.mkdir()
            def run(checks,name,command,*args,**kwargs):
                if name=='toolchain':return 'go'+(gate.ROOT/'.go-version').read_text().strip()+'\nlinux\namd64\noff\n'
                if name.startswith('download-'):
                    (checks.output/(name+'.log')).write_text('unknown revision')
                    checks.steps.append({'name':name,'status':'failed'})
                    raise RuntimeError('unknown revision')
                self.fail('validation must not run after failed downloads')
            with patch.object(gate.Checks,'run',run):result=gate.consume(root/'cold',output,'v0.1.0-rc.1','v0.1.0-rc.1','v0.8.0')
            self.assertEqual(result['status'],'failed')
            self.assertEqual(len(result['steps']),3)
            self.assertEqual(len(list(output.glob('download-*.log'))),3)
            self.assertTrue((output/'summary.json').is_file())

    def test_existing_output_is_never_reused(self):
        with tempfile.TemporaryDirectory() as path:
            with self.assertRaises(FileExistsError):gate.execute('v0.1.0-rc.1','v0.1.0-rc.1',Path(path))

    def test_tag_and_authenticated_download_must_identify_same_commit(self):
        tag = 'adapters/sctp/v0.1.0-rc.1'
        commit = 'a' * 40
        origin = {'VCS':'git','URL':'https://github.com/skill215/diameterkit','Hash':commit,'Ref':'refs/tags/'+tag}
        gate.validate_origin({'Origin':origin},tag,commit)
        # Go's Origin.Ref is optional (codehost.Origin); exact version plus the
        # independently resolved tag commit still binds the downloaded source.
        without_ref = dict(origin)
        without_ref.pop('Ref')
        gate.validate_origin({'Origin':without_ref},tag,commit)
        for change in ({'Hash':'b'*40},{'URL':'https://example.com/fake'},{'Ref':'refs/heads/main'},{'VCS':''}):
            with self.assertRaises(RuntimeError):gate.validate_origin({'Origin':{**origin,**change}},tag,commit)
        with self.assertRaises(RuntimeError):gate.validate_origin({},tag,commit)

if __name__=='__main__':unittest.main()
