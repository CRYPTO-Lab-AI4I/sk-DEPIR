"""Add only current-style progress observation to the exact old resident schedule.

This is a measurement-only source patch, never a production implementation.
Keep its diff and pristine/modified hashes in the run evidence. Both binaries
run with logging enabled; results include observer cost and are not pristine
unlogged historical timings. No arithmetic, loop ordering or allocation changes.
"""
from pathlib import Path
import shutil


def replace_once(text,old,new):
    assert text.count(old)==1,old[:120]
    return text.replace(old,new,1)


def instrument(baseline,current):
    baseline,current=Path(baseline),Path(current)
    cpp=baseline/'encoding/liftrs-cpp'
    shutil.copyfile(current/'encoding/liftrs-cpp/progress.hpp',cpp/'progress.hpp')
    p=cpp/'matrix.hpp';text=p.read_text()
    text=replace_once(text,'#include "transform.hpp"','#include "transform.hpp"\n#include "progress.hpp"')
    text=replace_once(text,'template<class F> void parallel_jobs(size_t jobs,unsigned workers,F fn){',
                     'template<class F> void parallel_jobs(size_t jobs,unsigned workers,F fn){\n  WorkProgress progress("parallel-blocks",jobs);')
    text=replace_once(text,'for(size_t j=0;j<jobs;j++)fn(j);','for(size_t j=0;j<jobs;j++){fn(j);progress.advance();}')
    text=replace_once(text,'if(j>=jobs)break;fn(j);','if(j>=jobs)break;fn(j);progress.advance();')
    p.write_text(text)
    p=cpp/'c_api.cpp';text=p.read_text()
    text=replace_once(text,'    auto measured = [&](unsigned stage, auto fn) {', '''    liftrs::execution::EncodingProgress progress;
    unsigned buffer_stage=0,product_stage=0,fft_stage=0;
    const char* fft_phases[]={"prefix-p-forward","prefix-message-forward","prefix-inverse","full-forward","full-inverse"};
    auto measured = [&](unsigned stage, auto fn) {''')
    text=replace_once(text,'''      fn();
      stats[stage] += ''','''      const char* label=stage==3?(buffer_stage++==0?"prepare-prefix":"promote-prefix"):
        stage==4?(product_stage++==0?"prefix-product":"spectrum-product"):
        stage==5?"canonical-assembly":stage==6?"compaction":"packed-output";
      progress.run(label,fn);
      stats[stage] += ''')
    text=replace_once(text,'''    auto fft = [&](bool prefix, uint32_t* data, bool inverse) {
      uint64_t stages[3] = {};''','''    auto fft = [&](bool prefix, uint32_t* data, bool inverse) {
      progress.run(fft_phases[fft_stage++],[&]{
      uint64_t stages[3] = {};''')
    text=replace_once(text,'''      for (unsigned j = 0; j < 3; j++) stats[j] += stages[j];
    };''','''      for (unsigned j = 0; j < 3; j++) stats[j] += stages[j];
      });
    };''')
    p.write_text(text)
