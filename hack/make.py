#!/usr/bin/env python


# http://stackoverflow.com/a/14050282
def check_antipackage():
    from sys import version_info
    sys_version = version_info[:2]
    found = True
    if sys_version < (3, 0):
        # 'python 2'
        from pkgutil import find_loader
        found = find_loader('antipackage') is not None
    elif sys_version <= (3, 3):
        # 'python <= 3.3'
        from importlib import find_loader
        found = find_loader('antipackage') is not None
    else:
        # 'python >= 3.4'
        from importlib import util
        found = util.find_spec('antipackage') is not None
    if not found:
        print('Install missing package "antipackage"')
        print('Example: pip install git+https://github.com/ellisonbg/antipackage.git#egg=antipackage')
        from sys import exit
        exit(1)
check_antipackage()

# ref: https://github.com/ellisonbg/antipackage
import antipackage
from github.appscode.libbuild import libbuild

import os
import os.path
import shutil
import subprocess
import sys
import tempfile
from os.path import expandvars

# Debian package
# https://gist.github.com/rcrowley/3728417
libbuild.REPO_ROOT = expandvars('$GOPATH') + '/src/appscode.com/krank'
BUILD_METADATA = libbuild.metadata(libbuild.REPO_ROOT)
libbuild.BIN_MATRIX = {
    'start-kubernetes': {
        'type': 'go',
        'go_version': True,
        'release': True,
        'distro': {
            'linux': [
                'amd64'
            ]
        }
    }
}
libbuild.BUCKET_MATRIX = {
    'prod': {
        'gs://appscode-asia': '',
        'gs://appscode-eu': '',
        'gs://appscode-us': '',
        's3://appscode-frankfurt': 'eu-central-1',
        's3://appscode-ireland': 'eu-west-1',
        's3://appscode-london': 'eu-west-2',
        's3://appscode-montreal': 'ca-central-1',
        's3://appscode-mumbai': 'ap-south-1',
        's3://appscode-norcal': 'us-west-1',
        's3://appscode-ohio': 'us-east-2',
        's3://appscode-oregon': 'us-west-2',
        's3://appscode-saopaulo': 'sa-east-1',
        's3://appscode-seoul': 'ap-northeast-2',
        's3://appscode-singapore': 'ap-southeast-1',
        's3://appscode-sydney': 'ap-southeast-2',
        's3://appscode-tokyo': 'ap-northeast-1',
        's3://appscode-virginia': 'us-east-1',
    },
    'dev': {
        'gs://appscode-dev': '',
        's3://appscode-dev': 'us-east-1',
    }
}


def call(cmd, stdin=None, cwd=libbuild.REPO_ROOT):
    print(cmd)
    return subprocess.call([expandvars(cmd)], shell=True, stdin=stdin, cwd=cwd)


def die(status):
    if status:
        sys.exit(status)


def check_output(cmd, stdin=None, cwd=libbuild.REPO_ROOT):
    print(cmd)
    return subprocess.check_output([expandvars(cmd)], shell=True, stdin=stdin, cwd=cwd)


def version():
    # json.dump(BUILD_METADATA, sys.stdout, sort_keys=True, indent=2)
    for k in sorted(BUILD_METADATA):
        print(k + '=' + BUILD_METADATA[k])


def fmt():
    libbuild.ungroup_go_imports('pkg', 'cmd')
    die(call('goimports -w pkg cmd'))
    call('gofmt -s -w pkg cmd')


def vet():
    call('go vet ./pkg/... ./cmd/...')


def deps():
    libbuild.deps()


def gen_protos():
    # Generate protos
    die(call('./hack/gen.sh', cwd=libbuild.REPO_ROOT + '/_proto'))
    #Move generated go files to api.
    call('rm -rf pkg/apis')
    call('mkdir -p pkg/apis')
    call("find . -name '*.go' | cpio -pdm "+libbuild.REPO_ROOT+"/pkg/apis", cwd=libbuild.REPO_ROOT + '/_proto')
    call("find . -type f -name '*.go' -delete", cwd=libbuild.REPO_ROOT + '/_proto')


def gen_assets():
    die(call('go-bindata -ignore=\\.go -ignore=\\.DS_Store -mode=0644 -modtime=1453795200 -o bindata.go -pkg templates ./...', cwd=libbuild.REPO_ROOT + '/pkg/templates'))

def gen_extpoints():
    die(call('go generate cmd/start-kubernetes/main.go'))


def gen():
    gen_assets()
    gen_extpoints()


def build_cmd(name):
    cfg = libbuild.BIN_MATRIX[name]
    if cfg['type'] == 'go':
        if 'distro' in cfg.keys():
            for goos, archs in cfg['distro'].items():
                for goarch in archs:
                    libbuild.go_build(name, goos, goarch, main='cmd/{}/*.go'.format(name))
        else:
            libbuild.go_build(name, libbuild.GOHOSTOS, libbuild.GOHOSTARCH, main='cmd/{}/*.go'.format(name))


def build_cmds():
    gen()
    fmt()
    for name in libbuild.BIN_MATRIX.keys():
        build_cmd(name)


def build(name=None):
    if name:
        cfg = libbuild.BIN_MATRIX[name]
        if cfg['type'] == 'go':
            gen()
            fmt()
            build_cmd(name)
    else:
        build_cmds()


def push(name=None):
    if name:
        bindir = libbuild.REPO_ROOT + '/dist/' + name
        push_bin(bindir)
    else:
        dist = libbuild.REPO_ROOT + '/dist'
        for name in os.listdir(dist):
            d = dist + '/' + name
            if os.path.isdir(d):
                push_bin(d)


def push_bin(bindir):
    call('rm -f *.md5', cwd=bindir)
    call('rm -f *.sha1', cwd=bindir)
    for f in os.listdir(bindir):
        if os.path.isfile(bindir + '/' + f):
            libbuild.upload_to_cloud(bindir, f, BUILD_METADATA['version'])


def import_kubernetes(version):
    if version.startswith('v'):
        version = version[1:]
    dist = libbuild.REPO_ROOT + '/dist'
    if not os.path.exists(dist):
        os.makedirs(dist)
    if version.startswith('1.2.') or version.startswith('1.3.') or version in ['1.4.0', '1.4.1', '1.4.2', '1.4.3']:
        die(call('wget https://storage.googleapis.com/kubernetes-release/release/v{}/kubernetes.tar.gz'.format(version),
                 cwd=dist))
        die(call('tar zxvf kubernetes.tar.gz', cwd=dist))
        import_old_k8s_server(version)
        import_old_kubectl(version)
        call('rm -rf kubernetes kubernetes.tar.gz {}'.format(version), cwd=dist)
    else:
        import_new_k8s_server(version)
        import_new_kube_clis(version)


# import from old format, from a single archive
def import_old_k8s_server(version, bucket_prefix=None):
    dist = libbuild.REPO_ROOT + '/dist'
    call('mv kubernetes/server/kubernetes-server-linux-amd64.tar.gz .', cwd=dist)
    libbuild.write_checksum(dist, 'kubernetes-server-linux-amd64.tar.gz')
    for bucket, region in libbuild.BUCKET_MATRIX.get(libbuild.ENV, libbuild.BUCKET_MATRIX['dev']).items():
        if bucket_prefix and not bucket.startswith(bucket_prefix):
            continue
        dst = "{bucket}/binaries/kubernetes-server/{version}/kubernetes-server-linux-amd64.tar.gz".format(
            bucket=bucket,
            version=version
        )
        if bucket.startswith('gs://'):
            libbuild.upload_to_gcs(dist, 'kubernetes-server-linux-amd64.tar.gz', dst, True)
        elif bucket.startswith('s3://'):
            libbuild.upload_to_s3(dist, 'kubernetes-server-linux-amd64.tar.gz', dst, region, True)


# import from old format, from a single archive
def import_old_kubectl(version):
    dir = libbuild.REPO_ROOT + '/dist/kubernetes/platforms'

    for root, dirs, files in os.walk(dir):
        for name in files:
            if name.startswith('kubectl'):
                _, goos, goarch = root.rsplit('/', 2)
                print(goos, goarch, root + '/' + name)
                file = root + '/' + name
                call('gsutil cp -r {file} gs://appscode-cdn/binaries/kubectl/{version}/kubectl-{goos}-{goarch}{ext}'.format(
                    file=file,
                    version=version,
                    goos=goos,
                    goarch=goarch,
                    ext='.exe' if goos == 'windows' else ''
                ))
                call('gsutil acl ch -u AllUsers:R -r gs://appscode-cdn/binaries/kubectl/{}'.format(version))
    lf = libbuild.REPO_ROOT + '/dist/kubernetes/platforms/latest.txt'
    libbuild.write_file(lf, version)
    call("gsutil cp {0} gs://appscode-cdn/binaries/kubectl/latest.txt".format(lf))
    call('gsutil acl ch -u AllUsers:R -r gs://appscode-cdn/binaries/kubectl/latest.txt')


# import from new format
def import_new_k8s_server(version, bucket_prefix=None):
    d = libbuild.REPO_ROOT + '/dist/kubernetes-server'
    die(call('mkdir -p {}'.format(d)))
    name = 'kubernetes-server-linux-amd64.tar.gz'
    die(call('wget https://dl.k8s.io/v{}/{}'.format(version, name), cwd=d))
    libbuild.write_checksum(d, name)
    for bucket, region in libbuild.BUCKET_MATRIX.get(libbuild.ENV, libbuild.BUCKET_MATRIX['dev']).items():
        if bucket_prefix and not bucket.startswith(bucket_prefix):
            continue
        dst = "{bucket}/binaries/kubernetes-server/{version}/{name}".format(
            bucket=bucket,
            version=version,
            name=name
        )
        if bucket.startswith('gs://'):
            libbuild.upload_to_gcs(d, name, dst, True)
        elif bucket.startswith('s3://'):
            libbuild.upload_to_s3(d, name, dst, region, True)
    die(call('rm -rf {}'.format(d)))


# import from new format
def import_new_kube_clis(version):
    distros = {
        'darwin': ['386', 'amd64'],
        'linux': ['386', 'amd64', 'arm', 'arm64'],
        'windows': ['386', 'amd64']
    }
    bins = []
    for goos, archs in distros.items():
        for goarch in archs:
            d = libbuild.REPO_ROOT + '/dist/kubernetes-client'
            die(call('mkdir -p {}'.format(d)))
            archive = 'kubernetes-client-{}-{}.tar.gz'.format(goos, goarch)
            die(call('wget https://dl.k8s.io/v{}/{}'.format(version, archive), cwd=d))
            die(call('tar zxvf {}'.format(archive), cwd=d))
            print()
            for root, dirs, files in os.walk(d + '/kubernetes/client/bin'):
                for name in files:
                    if name.endswith('.exe'):
                        name = name[:-len('.exe')]
                    bins.append(name)
                    call('gsutil cp -r {file}{ext} gs://appscode-cdn/binaries/{name}/{version}/{name}-{os}-{arch}{ext}'.format(
                        file=root + '/' + name,
                        version=version,
                        name=name,
                        os=goos,
                        arch=goarch,
                        ext='.exe' if goos == 'windows' else ''
                    ))
                    call('gsutil acl ch -u AllUsers:R -r gs://appscode-cdn/binaries/{}/{}'.format(name, version))
            die(call('rm -rf {}'.format(d)))
    for name in frozenset(bins):
        lf = libbuild.REPO_ROOT + '/dist/latest.txt'
        libbuild.write_file(lf, version)
        call("gsutil cp {0} gs://appscode-cdn/binaries/{1}/latest.txt".format(lf, name))
        call('gsutil acl ch -u AllUsers:R -r gs://appscode-cdn/binaries/{0}/latest.txt'.format(name))
        die(call('rm -f {}'.format(lf)))


def import_kubernetes_salt(version, bucket_prefix=None):
    call('git clean -xfd', cwd=expandvars('$GOPATH/src/k8s.io/kubernetes'))
    die(call('git checkout ' + version, cwd=expandvars('$GOPATH/src/k8s.io/kubernetes')))
    die(call('git pull origin ' + version, cwd=expandvars('$GOPATH/src/k8s.io/kubernetes')))

    tmp = tempfile.mkdtemp()
    shutil.copytree(expandvars('$GOPATH/src/k8s.io/kubernetes/cluster/saltbase'), tmp + '/kubernetes/saltbase')
    addons = expandvars('$GOPATH/src/k8s.io/kubernetes/cluster/addons')
    for dirname, dirnames, filenames in os.walk(addons):
        for filename in filenames:
            if ".yaml" in filename:
                src = os.path.join(dirname, filename)
                dst = os.path.join(tmp + '/kubernetes/saltbase/salt/kube-addons' + dirname[len(addons):], filename)
                # print src + ' -> ' + dst
                dir = os.path.dirname(dst)
                if not os.path.exists(dir):
                    os.makedirs(dir)
                shutil.copyfile(src, dst)
    call('tar -czf ' + tmp + '/kubernetes-salt.tar.gz kubernetes', cwd=tmp)
    dist = libbuild.REPO_ROOT + '/dist/kubernetes-salt/' + version
    if not os.path.exists(dist):
        os.makedirs(dist)
    shutil.copyfile(tmp + '/kubernetes-salt.tar.gz', dist + '/kubernetes-salt.tar.gz')
    shutil.rmtree(tmp, ignore_errors=True)

    libbuild.write_checksum(dist, 'kubernetes-salt.tar.gz')
    for bucket, region in libbuild.BUCKET_MATRIX.get(libbuild.ENV, libbuild.BUCKET_MATRIX['dev']).items():
        if bucket_prefix and not bucket.startswith(bucket_prefix):
            continue
        dst = "{bucket}/binaries/kubernetes-salt/{version}/kubernetes-salt.tar.gz".format(
            bucket=bucket,
            version=version
        )
        if bucket.startswith('gs://'):
            libbuild.upload_to_gcs(dist, 'kubernetes-salt.tar.gz', dst, True)
        elif bucket.startswith('s3://'):
            libbuild.upload_to_s3(dist, 'kubernetes-salt.tar.gz', dst, region, True)
    call('rm -rf kubernetes-salt', cwd=libbuild.REPO_ROOT + '/dist')


def make_public(bucket_prefix=None):
    for bucket in libbuild.BUCKET_MATRIX.get(libbuild.ENV, libbuild.BUCKET_MATRIX['dev']):
        if bucket_prefix and not bucket.startswith(bucket_prefix):
            continue
        if bucket.startswith('gs://'):
            call("gsutil acl ch -u AllUsers:R -r {bucket}/binaries/kubernetes-server".format(bucket=bucket))
        if bucket.startswith('gs://'):
            call("gsutil acl ch -u AllUsers:R -r {bucket}/binaries/kubectl".format(bucket=bucket))
        if bucket.startswith('gs://'):
            call("gsutil acl ch -u AllUsers:R -r {bucket}/binaries/appctl".format(bucket=bucket))
        for name in libbuild.BIN_MATRIX.keys():
            if libbuild.BIN_MATRIX[name].get('release', False):
                path = "{bucket}/binaries/{name}".format(
                    bucket=bucket,
                    name=name,
                )
                if bucket.startswith('gs://'):
                    call("gsutil acl ch -u AllUsers:R -r {0}".format(path))
                if bucket.startswith('s3://'):
                    print("**********************************")
                    print('make_public command is not supprted for S3. Please make {} public manually.'.format(path))
                    print("**********************************")
                    # http://docs.aws.amazon.com/cli/latest/reference/s3api/put-object-acl.html
                    # call("aws s3api put-object-acl --acl public-read {0}".format(cloud_dst), cwd=dir)


def delete_kubernetes_salt(version, bucket_prefix=None):
    _delete_cloud_binary('kubernetes-salt', version, bucket_prefix)
    # _delete_cloud_binary('start-kubernetes', version, bucket_prefix)


def delete_kubernetes(version):
    if version.startswith('v'):
        version = version[1:]
    delete_k8s_server(version)
    delete_kubectl(version)


def delete_k8s_server(version, bucket_prefix=None):
    _delete_cloud_binary('kubernetes-server', version, bucket_prefix)


def _delete_cloud_binary(binary, version, bucket_prefix=None):
    for bucket, region in libbuild.BUCKET_MATRIX.get(libbuild.ENV, libbuild.BUCKET_MATRIX['dev']).items():
        if bucket_prefix and not bucket.startswith(bucket_prefix):
            continue
        path = "{bucket}/binaries/{binary}/{version}/".format(
            bucket=bucket,
            binary=binary,
            version=version
        )
        if bucket.startswith('gs://'):
            call("gsutil rm -r {0}".format(path))
        elif bucket.startswith('s3://'):
            opt_region = ''
            if region:
                opt_region = '--region ' + region
            call("aws s3 rm {0} {1} --recursive".format(opt_region, path))


def delete_kubectl(version):
    call('gsutil rm -r gs://appscode-cdn/binaries/kubectl/{0}/'.format(version))


def default():
    gen()
    fmt()
    die(call('GO15VENDOREXPERIMENT=1 ' + libbuild.GOC + ' install ./pkg/... ./cmd/...'))


def install():
    die(call('GO15VENDOREXPERIMENT=1 ' + libbuild.GOC + ' install ./pkg/... ./cmd/...'))


if __name__ == "__main__":
    if len(sys.argv) > 1:
        # http://stackoverflow.com/a/834451
        # http://stackoverflow.com/a/817296
        globals()[sys.argv[1]](*sys.argv[2:])
    else:
        default()
