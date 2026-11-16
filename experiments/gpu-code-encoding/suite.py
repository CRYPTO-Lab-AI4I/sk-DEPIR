"""Client additions to the existing bounded persistent GPU regression controller."""


def run_suite(context):
    run = context['run']
    source, cache, out = (context[key] for key in ('source', 'cache', 'out'))
    env, go, library = (context[key] for key in ('env', 'go', 'library'))
    binary, fixtures, checker = cache/'code.test', cache/'code-fixtures', cache/'code-check'
    package = source/'encoding/liftrs-shared'
    run('code-build', [go, 'test', '-tags', 'liftrs_cpp,liftrs_cuda', '-c',
                      '-o', binary, './encoding/liftrs-shared'], source)
    env.update(PIR_GPU_PREPARATION_LIBRARY=str(library),
               PIR_GPU_CODE_FIXTURES=str(fixtures), LIFTRS_CPU_PROGRESS='1')
    run('code-fixtures', [binary, '-test.run=^TestCUDACodeNativeFixtures$', '-test.v',
                         '-test.cpuprofile='+str(out/'code-fixtures.pprof')], package)
    # Recheck real s17 field transforms/products and cross-width peer leases.
    # 64MiB/card prevents the negative full-enumeration test allocating a bitmap.
    env.update(PIR_GPU_PREPARATION_DEVICES='0,1,2,3',
               PIR_GPU_PREPARATION_CAP=str(64 << 20))
    run('code-wide-kernels', [binary,
        '-test.run=^TestGPUPreparation(Wide|PeerLifetime)$', '-test.v',
        '-test.cpuprofile='+str(out/'code-wide-kernels.pprof')], package)
    env.pop('PIR_GPU_PREPARATION_DEVICES')
    env.pop('PIR_GPU_PREPARATION_CAP')
    run('code-small', [binary, '-test.run=^TestCUDACode(Small|Resources)$', '-test.v',
                      '-test.cpuprofile='+str(out/'code-small.pprof')], package)
    run('code-check-build', ['nvcc', '-std=c++17', '-O2', '-lineinfo', '-arch=native',
        '-Xcompiler=-pthread', 'experiments/gpu-code-encoding/check.cu', '-o', checker], source)
    run('code-check', [checker, fixtures/'native.bin'], source)
    run('code-memcheck', ['compute-sanitizer', '--tool', 'memcheck', '--leak-check',
        'full', '--error-exitcode', '99', checker, fixtures/'native.bin', '--small'], source)
    log = (out/'code-memcheck.log').read_text()
    assert 'ERROR SUMMARY: 0 errors' in log and 'LEAK SUMMARY: 0 bytes' in log
    env.pop('PIR_GPU_CODE_FIXTURES')
    env.pop('PIR_GPU_PREPARATION_LIBRARY')
