// Read an existing raw support archive once. Histograms use the production
// closed-orbit ownership rule; no support is inferred or generated here.
#include <openssl/evp.h>
#include <algorithm>
#include <array>
#include <chrono>
#include <cstdint>
#include <ctime>
#include <fstream>
#include <iomanip>
#include <iostream>
#include <memory>
#include <stdexcept>
#include <string>
#include <vector>
#include <sys/resource.h>

using Clock = std::chrono::steady_clock;
void progress(const char* event, uint64_t done, uint64_t total, Clock::time_point began) {
    rusage usage{};
    getrusage(RUSAGE_SELF, &usage);
    auto now = std::time(nullptr);
    std::tm utc{};
    gmtime_r(&now, &utc);
    char stamp[32];
    std::strftime(stamp, sizeof(stamp), "%Y-%m-%dT%H:%M:%SZ", &utc);
    std::cerr << "{\"utc\":\"" << stamp << "\",\"phase\":\"support-owner-scan\",\"event\":\""
              << event << "\",\"completed\":" << done << ",\"total\":" << total
              << ",\"unit\":\"indices\",\"peak_rss_kib\":" << usage.ru_maxrss << ",\"seconds\":"
              << std::chrono::duration<double>(Clock::now()-began).count() << "}" << std::endl;
}
int main(int argc, char** argv) {
    try {
        if (argc != 7) throw std::runtime_error("arguments: s word-bytes count sha256 support-path output-json");
        unsigned s = std::stoul(argv[1]), width = std::stoul(argv[2]);
        uint64_t expected = std::stoull(argv[3]);
        std::string expected_hash = argv[4];
        if (s < 2 || s > 18 || (width != 4 && width != 8) || expected_hash.size() != 64)
            throw std::runtime_error("invalid scan parameters");
        uint64_t q = uint64_t(1) << s, n = q*q;
        if (!expected || expected >= n) throw std::runtime_error("invalid support count");
        std::ifstream input(argv[5], std::ios::binary);
        if (!input) throw std::runtime_error("cannot open support");
        auto began = Clock::now(), last = began;
        using Digest = std::unique_ptr<EVP_MD_CTX, decltype(&EVP_MD_CTX_free)>;
        Digest digest(EVP_MD_CTX_new(), EVP_MD_CTX_free);
        if (!digest || EVP_DigestInit_ex(digest.get(), EVP_sha256(), nullptr) != 1)
            throw std::runtime_error("SHA256 initialization");
        std::vector<unsigned char> buffer(8 << 20);
        std::array<uint64_t, 64> counts{};
        uint64_t done = 0, previous = 0;
        progress("started", done, expected, began);
        while (input) {
            input.read(reinterpret_cast<char*>(buffer.data()), buffer.size());
            size_t bytes = input.gcount();
            if (bytes % width) throw std::runtime_error("truncated support word");
            if (EVP_DigestUpdate(digest.get(), buffer.data(), bytes) != 1)
                throw std::runtime_error("SHA256 update");
            for (size_t offset = 0; offset < bytes; offset += width) {
                uint64_t v = 0;
                for (unsigned j = 0; j < width; j++) v |= uint64_t(buffer[offset+j]) << (8*j);
                if (v >= n-1 || (done && v <= previous) || done >= expected)
                    throw std::runtime_error("support order/range/count");
                // q is a multiple of 64 for s>=6. For smaller regression
                // shapes use the general formula, which also fits uint64.
                uint64_t orbit = std::min(v, v*q % (n-1));
                counts[orbit % 64]++;
                previous = v;
                done++;
            }
            if (Clock::now()-last >= std::chrono::seconds(30)) {
                progress("progress", done, expected, began);
                last = Clock::now();
            }
        }
        if (!input.eof() || done != expected) throw std::runtime_error("support read/count");
        unsigned char hash[EVP_MAX_MD_SIZE]; unsigned hash_bytes = 0;
        if (EVP_DigestFinal_ex(digest.get(), hash, &hash_bytes) != 1 || hash_bytes != 32)
            throw std::runtime_error("SHA256 finalization");
        std::string actual_hash;
        const char* hex = "0123456789abcdef";
        for (unsigned j = 0; j < hash_bytes; j++) {
            actual_hash += hex[hash[j] >> 4]; actual_hash += hex[hash[j] & 15];
        }
        if (actual_hash != expected_hash) throw std::runtime_error("support SHA256 mismatch");
        std::ofstream out(argv[6], std::ios::out | std::ios::trunc);
        if (!out) throw std::runtime_error("cannot open result");
        out << "{\"s\":" << s << ",\"count\":" << done << ",\"bytes\":" << done*width
            << ",\"sha256\":\"" << actual_hash << "\",\"seconds\":"
            << std::setprecision(12) << std::chrono::duration<double>(Clock::now()-began).count()
            << ",\"owners\":{";
        for (unsigned ranks = 2; ranks <= 64; ranks *= 2) {
            if (ranks != 2) out << ',';
            out << '"' << ranks << "\":[";
            for (unsigned r = 0; r < ranks; r++) {
                if (r) out << ',';
                uint64_t count = 0;
                for (unsigned k = r; k < 64; k += ranks) count += counts[k];
                out << count;
            }
            out << ']';
        }
        out << "}}\n";
        out.close();
        if (!out) throw std::runtime_error("cannot flush result");
        progress("completed", done, expected, began);
    } catch (const std::exception& e) {
        std::cerr << "support-owner-scan error: " << e.what() << std::endl;
        return 1;
    }
}
