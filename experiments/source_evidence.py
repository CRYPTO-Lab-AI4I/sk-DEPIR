"""Resolve recorded source hashes across path renames and historical dirty runs.

Current files are checked first, then recorded revisions/history. Historical
manifests remain immutable; matching a later committed blob recovers the exact
measured bytes without pretending the measured working tree was clean.
"""
import hashlib,subprocess
from pathlib import Path

def verify_sources(root,sources,revision,snapshots=None):
    root=Path(root);snapshots=snapshots or {};resolved={}
    for name,want in sources.items():
        choices=[Path(snapshots[name])] if name in snapshots else []
        choices+=[root/name,root/name.replace('encoding/liftrs/','encoding/liftrs-go/')]
        for path in choices:
            if path.is_file() and hashlib.sha256(path.read_bytes()).hexdigest()==want:
                resolved[name]=str(path.relative_to(root));break
        else:
            revisions=[revision]
            history=subprocess.run(['git','log','--all','--format=%H','--',name],cwd=root,capture_output=True,text=True,check=True).stdout.splitlines()
            for commit in dict.fromkeys(revisions+history):
                blob=subprocess.run(['git','show',commit+':'+name],cwd=root,capture_output=True)
                if blob.returncode==0 and hashlib.sha256(blob.stdout).hexdigest()==want:
                    resolved[name]=commit+':'+name;break
            else:raise AssertionError('Recorded source unavailable: '+name)
    return resolved
