"""Linux file-cache residency without reading/faulting the file into process RSS."""
import ctypes
import os
from pathlib import Path

LIBC = ctypes.CDLL(None, use_errno=True)
LIBC.mmap.restype = ctypes.c_void_p
LIBC.mmap.argtypes = [ctypes.c_void_p, ctypes.c_size_t, ctypes.c_int, ctypes.c_int, ctypes.c_int, ctypes.c_long]
LIBC.munmap.argtypes = [ctypes.c_void_p, ctypes.c_size_t]
LIBC.mincore.argtypes = [ctypes.c_void_p, ctypes.c_size_t, ctypes.c_void_p]
PAGE = os.sysconf('SC_PAGE_SIZE')
BITS = bytes(i & 1 for i in range(256))


class FileCache:
    def __init__(self):
        self.maps = {}

    def close(self):
        for size, fd, address in self.maps.values():
            LIBC.munmap(address, size)
            os.close(fd)
        self.maps.clear()

    def sample(self, paths):
        result = {}
        for path in sorted(set(str(Path(p).resolve()) for p in paths)):
            try:
                size = Path(path).stat().st_size
            except FileNotFoundError:
                continue
            if not size:
                continue
            old = self.maps.get(path)
            if old and old[0] != size:
                LIBC.munmap(old[2], old[0]); os.close(old[1]); del self.maps[path]
            if path not in self.maps:
                try:
                    fd = os.open(path, os.O_RDONLY)
                except FileNotFoundError:
                    continue  # A packed output may have just been atomically renamed.
                # PROT_NONE/MAP_SHARED reserves address space only; no contents read.
                address = LIBC.mmap(None, size, 0, 1, fd, 0)
                if address == ctypes.c_void_p(-1).value:
                    os.close(fd); raise OSError(ctypes.get_errno(), 'cache observation mmap')
                self.maps[path] = size, fd, address
            pages = (size + PAGE - 1) // PAGE
            vector = (ctypes.c_ubyte * pages)()
            if LIBC.mincore(self.maps[path][2], size, vector):
                raise OSError(ctypes.get_errno(), 'cache observation mincore')
            result[path] = bytes(vector).translate(BITS).count(b'\x01') * PAGE
        return result


def job_cgroup(job):
    for line in Path('/proc/self/cgroup').read_text().splitlines():
        if not line.startswith('0::'):
            continue
        relative = line[3:]
        if str(job) not in relative:
            continue  # Never treat a whole user/node scope as this job's charge.
        base = Path('/sys/fs/cgroup') / relative.lstrip('/')
        if not (base/'memory.current').exists():
            continue
        stat = dict(row.split() for row in (base/'memory.stat').read_text().splitlines())
        return {'path': str(base), 'current_bytes': int((base/'memory.current').read_text()),
                'anon_bytes': int(stat.get('anon', 0)), 'file_bytes': int(stat.get('file', 0)),
                'peak_bytes': int((base/'memory.peak').read_text()) if (base/'memory.peak').exists() else None}
    return None
