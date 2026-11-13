#pragma once
// Explicit disk-backed component. Files are mutable, exclusive caller-owned
// little-endian arrays. Bounded private buffers do not bound filesystem cache.
#include <sys/stat.h>
#include <fcntl.h>
extern "C" int liftrs_matrix_apply_file(const liftrs_plan* rows,const liftrs_plan* columns,int data_fd,int scratch_fd,uint64_t data_offset,uint64_t scratch_offset,size_t n,size_t tile,unsigned split,unsigned workers,int inverse,uint64_t* stats,char* error,size_t capacity){
 try{
  if(!rows||!columns||!stats||data_fd<0||scratch_fd<0||!workers||workers>64||(inverse!=0&&inverse!=1)||n<4||(n&(n-1))||tile<4||tile>(size_t(1)<<26))throw std::runtime_error("disk transform arguments");
  const auto& r=rows->transform;const auto& c=columns->transform;unsigned dim=0;for(size_t v=n;v>1;v>>=1)dim++;
  if(!split||split>=dim||n>r.f.q||r.terms.size()!=split+1||c.terms.size()!=dim-split||rows->taylor_terms.size()!=dim-split||r.f.q!=c.f.q||r.f.polynomial!=c.f.polynomial||r.f.delta!=c.f.delta)throw std::runtime_error("disk transform shape/field");
  size_t n1=size_t(1)<<split,n2=n/n1;if(tile<std::max(n1,n2))throw std::runtime_error("disk tile shorter than vector");
  for(size_t i=0;i<r.basis.size();i++)if(r.basis[i]!=(uint32_t(1)<<i))throw std::runtime_error("disk rows require canonical basis");
  struct stat data_info,scratch_info;if(fstat(data_fd,&data_info)||fstat(scratch_fd,&scratch_info)||!S_ISREG(data_info.st_mode)||!S_ISREG(scratch_info.st_mode)||(data_info.st_dev==scratch_info.st_dev&&data_info.st_ino==scratch_info.st_ino))throw std::runtime_error("disk files must be distinct regular files");
  for(auto fd:{data_fd,scratch_fd}){int flags=fcntl(fd,F_GETFL);if(flags<0||(flags&O_ACCMODE)!=O_RDWR)throw std::runtime_error("disk files require read/write access");}
  const uint64_t max_offset=uint64_t(1)<<63;
  if(data_offset%4||scratch_offset%4||data_offset>max_offset-1-4*uint64_t(n)||scratch_offset>max_offset-1-4*uint64_t(n)||uint64_t(data_info.st_size)<data_offset+4*uint64_t(n)||uint64_t(scratch_info.st_size)<scratch_offset+4*uint64_t(n))throw std::runtime_error("disk file offsets/length");
  std::fill(stats,stats+12,0);auto begin=std::chrono::steady_clock::now();tile=std::min(tile,n);std::vector<uint32_t> a(tile),b(tile);stats[9]=4*(a.capacity()+b.capacity())+65536;
  if(a.capacity()>tile||b.capacity()>tile)throw std::runtime_error("disk private buffer bound");
  auto ns=[](auto start){return uint64_t(std::chrono::duration_cast<std::chrono::nanoseconds>(std::chrono::steady_clock::now()-start).count());};
  auto read=[&](int fd,uint64_t offset,uint32_t* dst,size_t count){auto start=std::chrono::steady_clock::now();size_t done=0,bytes=count*4;auto* raw=reinterpret_cast<unsigned char*>(dst);
   while(done<bytes){ssize_t got=pread(fd,raw+done,bytes-done,off_t(offset+done));stats[7]++;if(got<0&&errno==EINTR)continue;if(got<=0)throw std::runtime_error("disk pread failed/short file");done+=size_t(got);stats[5]+=uint64_t(got);}
   for(size_t i=0;i<count;i++){size_t j=4*i;dst[i]=uint32_t(raw[j])|(uint32_t(raw[j+1])<<8)|(uint32_t(raw[j+2])<<16)|(uint32_t(raw[j+3])<<24);if(uint64_t(dst[i])>=r.f.q)throw std::runtime_error("disk value outside field");}stats[3]+=ns(start);
  };
  auto write=[&](int fd,uint64_t offset,const uint32_t* src,size_t count){auto start=std::chrono::steady_clock::now();unsigned char raw[65536];
   for(size_t first=0;first<count;){size_t words=std::min(count-first,size_t(16384));for(size_t i=0;i<words;i++){uint32_t v=src[first+i];for(unsigned j=0;j<4;j++)raw[4*i+j]=static_cast<unsigned char>(v>>(8*j));}
    size_t done=0,bytes=words*4;while(done<bytes){ssize_t sent=pwrite(fd,raw+done,bytes-done,off_t(offset+first*4+done));stats[8]++;if(sent<0&&errno==EINTR)continue;if(sent<=0)throw std::runtime_error("disk pwrite failed");done+=size_t(sent);stats[6]+=uint64_t(sent);}first+=words;
   }stats[4]+=ns(start);
  };
  auto measured=[&](unsigned phase,auto fn){auto start=std::chrono::steady_clock::now();fn();stats[phase]+=ns(start);};
  auto convert=[&]{for(unsigned step=0;step<dim-split;step++){
   unsigned k=inverse?step:dim-split-1-step;const auto& terms=rows->taylor_terms[k];if(terms.empty())continue;size_t half=size_t(1)<<(split+k),minimum=half,maximum=0;
   for(const auto& t:terms){if(!t.degree||(t.degree&(t.degree-1))||t.degree>half/2)throw std::runtime_error("disk Taylor dependency");minimum=std::min(minimum,size_t(t.degree));maximum=std::max(maximum,size_t(t.degree));}
   auto pass=[&](bool quotient){for(size_t block=0;block<n;block+=2*half){size_t first=quotient?0:minimum,last=quotient?maximum:half;
    for(size_t j=first;j<last;){size_t count=std::min(tile,last-j),target=block+(quotient?half:0)+j;read(data_fd,data_offset+4*target,a.data(),count);
     for(const auto& t:terms){size_t d=t.degree,lo=quotient?j:std::max(j,d),hi=quotient?std::min(j+count,d):j+count;if(lo>=hi)continue;
      size_t source=block+(quotient?2*half:half)+lo-d;read(data_fd,data_offset+4*source,b.data(),hi-lo);
      parallel_jobs((hi-lo+8191)/8192,workers,[&](size_t job){for(size_t i=job*8192;i<std::min(hi-lo,(job+1)*8192);i++)a[lo-j+i]^=t.product.mul(r.f,b[i]);});
     }write(data_fd,data_offset+4*target,a.data(),count);j+=count;
    }
   }};if(inverse){pass(false);pass(true);}else{pass(true);pass(false);}
  }};
  auto transpose_file=[&](int src,uint64_t src_offset,int dst,uint64_t dst_offset,size_t height,size_t width){
   size_t block_rows=1;while(block_rows*2<=height&&(block_rows*2)<=tile/(block_rows*2))block_rows*=2;size_t block_cols=std::min(width,tile/block_rows);
   for(size_t row=0;row<height;row+=block_rows)for(size_t col=0;col<width;col+=block_cols){size_t nr=std::min(block_rows,height-row),nc=std::min(block_cols,width-col);
    for(size_t i=0;i<nr;i++)read(src,src_offset+4*((row+i)*width+col),a.data()+i*nc,nc);
    transpose(a.data(),b.data(),nr,nc,workers);
    for(size_t j=0;j<nc;j++)write(dst,dst_offset+4*((col+j)*height+row),b.data()+j*nr,nr);
   }stats[11]++;
  };
  auto vectors=[&](int fd,uint64_t offset,size_t length,size_t total,const liftrs::Transform& plan,unsigned level,bool affine){
   size_t batch=tile/length;for(size_t first=0;first<total;){size_t count=std::min(batch,total-first),words=count*length;read(fd,offset+4*first*length,a.data(),words);
    parallel_jobs(count,workers,[&](size_t j){plan.recurse(a.data()+j*length,length,level,affine?uint32_t((first+j)*length):0,inverse!=0);});write(fd,offset+4*first*length,a.data(),words);first+=count;
   }
  };
  auto columns_phase=[&]{transpose_file(data_fd,data_offset,scratch_fd,scratch_offset,n2,n1);vectors(scratch_fd,scratch_offset,n2,n1,c,dim-split,false);transpose_file(scratch_fd,scratch_offset,data_fd,data_offset,n1,n2);};
  auto rows_phase=[&]{vectors(data_fd,data_offset,n1,n2,r,split,true);};
  if(!inverse){measured(0,convert);measured(1,columns_phase);measured(2,rows_phase);}else{measured(2,rows_phase);measured(1,columns_phase);measured(0,convert);}stats[10]=ns(begin);return 0;
 }catch(const std::exception& e){fail(error,capacity,e.what());return 1;}catch(...){fail(error,capacity,"disk transform failed");return 1;}
}

// Complete explicit disk schedule. The output alone stays in caller RAM;
// code and mutable domains stay in files. All field arithmetic is native.
extern "C" int liftrs_online_encode_disk(const liftrs_plan* pr,const liftrs_plan* pc,const liftrs_plan* fr,const liftrs_plan* fc,unsigned ps,unsigned fs,unsigned workers,uint32_t omega,int code_fd,uint64_t p_offset,uint64_t spectrum_offset,int data_fd,int scratch_fd,size_t tile,const uint32_t* message,size_t k,size_t n,size_t l,uint32_t* word,uint64_t* stats,char* error,size_t capacity){
 try{
  if(!pr||!pc||!fr||!fc||!message||!word||!stats||!k||k>n/2||l<2*k-1||l>n/2||(l&(l-1))||tile<4||tile>(size_t(1)<<26)||!workers||workers>64||code_fd<0||data_fd<0||scratch_fd<0||n!=fr->transform.f.q)throw std::runtime_error("disk online arguments");
  const auto& f=fr->transform.f;
  for(auto* plan:{pr,pc,fc})if(plan->transform.f.q!=f.q||plan->transform.f.s!=f.s||plan->transform.f.polynomial!=f.polynomial||plan->transform.f.delta!=f.delta)throw std::runtime_error("disk online fields");
  struct stat files[3];int descriptors[3]={code_fd,data_fd,scratch_fd};for(unsigned i=0;i<3;i++){
   if(fstat(descriptors[i],files+i)||!S_ISREG(files[i].st_mode))throw std::runtime_error("disk online regular files");
   for(unsigned j=0;j<i;j++)if(files[i].st_dev==files[j].st_dev&&files[i].st_ino==files[j].st_ino)throw std::runtime_error("disk online file alias");
   int flags=fcntl(descriptors[i],F_GETFL);if(flags<0||(i?((flags&O_ACCMODE)!=O_RDWR):((flags&O_ACCMODE)==O_WRONLY)))throw std::runtime_error("disk online descriptor access");
  }
  // Probe headers are25/33 bytes; code-word ranges need not be aligned.
  if(p_offset>uint64_t(INT64_MAX)-4*uint64_t(k+1)||spectrum_offset>uint64_t(INT64_MAX)-4*uint64_t(n)||uint64_t(files[0].st_size)<std::max(p_offset+4*uint64_t(k+1),spectrum_offset+4*uint64_t(n))||uint64_t(files[1].st_size)<4*uint64_t(n)||uint64_t(files[2].st_size)<4*uint64_t(n))throw std::runtime_error("disk online file length/offset");
  auto power=[&](size_t exponent){uint32_t result=1,x=omega;while(exponent){if(exponent&1)result=f.mul(result,x);x=f.mul(x,x);exponent>>=1;}return result;};
  if(!omega||uint64_t(omega)>=f.q||power(n-1)!=1)throw std::runtime_error("disk online generator");size_t order=n-1;for(size_t d=2;d<=order/d;d++){if(order%d)continue;if(power((n-1)/d)==1)throw std::runtime_error("disk online nonprimitive generator");while(order%d==0)order/=d;}if(order>1&&power((n-1)/order)==1)throw std::runtime_error("disk online nonprimitive generator");
  for(size_t i=0;i<k;i++)if(message[i]>=(1u<<f.s))throw std::runtime_error("disk online message outside field");
  std::fill(stats,stats+16,0);auto begin=std::chrono::steady_clock::now();tile=std::min(tile,n);std::vector<uint32_t>a(tile),b(tile);if(a.capacity()>tile||b.capacity()>tile)throw std::runtime_error("disk online tile capacity");
  // Outer product/assembly tiles remain live during nested transforms, which
  // own two more tiles. Include both pairs and conservative I/O stack staging.
  stats[12]=16*uint64_t(tile)+2*65536;stats[15]=8*uint64_t(n);
  auto ns=[](auto start){return uint64_t(std::chrono::duration_cast<std::chrono::nanoseconds>(std::chrono::steady_clock::now()-start).count());};
  auto read=[&](int fd,uint64_t offset,uint32_t* dst,size_t count,uint64_t limit){auto start=std::chrono::steady_clock::now();auto* raw=reinterpret_cast<unsigned char*>(dst);size_t done=0;
   while(done<count*4){ssize_t got=pread(fd,raw+done,count*4-done,off_t(offset+done));stats[10]++;if(got<0&&errno==EINTR)continue;if(got<=0)throw std::runtime_error("disk online pread failed");done+=size_t(got);stats[8]+=uint64_t(got);}
   for(size_t i=0;i<count;i++){size_t j=4*i;dst[i]=uint32_t(raw[j])|(uint32_t(raw[j+1])<<8)|(uint32_t(raw[j+2])<<16)|(uint32_t(raw[j+3])<<24);if(uint64_t(dst[i])>=limit)throw std::runtime_error("disk online file value outside field");}stats[6]+=ns(start);
  };
  auto write=[&](uint64_t offset,const uint32_t* src,size_t count){auto start=std::chrono::steady_clock::now();unsigned char raw[65536];for(size_t first=0;first<count;){size_t words=std::min(count-first,size_t(16384));for(size_t i=0;i<words;i++)for(unsigned j=0;j<4;j++)raw[4*i+j]=static_cast<unsigned char>(src[first+i]>>(8*j));size_t done=0;while(done<words*4){ssize_t sent=pwrite(data_fd,raw+done,words*4-done,off_t(offset+4*first+done));stats[11]++;if(sent<0&&errno==EINTR)continue;if(sent<=0)throw std::runtime_error("disk online pwrite failed");done+=size_t(sent);stats[9]+=uint64_t(sent);}first+=words;}stats[7]+=ns(start);};
  auto measured=[&](unsigned phase,auto fn){auto start=std::chrono::steady_clock::now();fn();stats[phase]+=ns(start);};
  auto zero=[&](size_t first,size_t count){std::fill(a.begin(),a.end(),0);while(count){size_t take=std::min(tile,count);write(4*uint64_t(first),a.data(),take);first+=take;count-=take;}};
  auto fft=[&](bool prefix,uint64_t offset,bool inverse){uint64_t st[12];char detail[512];if(liftrs_matrix_apply_file(prefix?pr:fr,prefix?pc:fc,data_fd,scratch_fd,offset,0,prefix?l:n,tile,prefix?ps:fs,workers,inverse,st,detail,sizeof(detail)))throw std::runtime_error(detail);
   for(unsigned i=0;i<3;i++)stats[i]+=st[i];for(unsigned i=3;i<9;i++)stats[i+3]+=st[i];stats[14]++;
  };
  auto product=[&](bool prefix){size_t count=prefix?l:n;for(size_t first=0;first<count;){size_t take=std::min(tile,count-first);read(data_fd,4*uint64_t(first),a.data(),take,f.q);read(prefix?data_fd:code_fd,(prefix?4*uint64_t(l):spectrum_offset)+4*uint64_t(first),b.data(),take,f.q);
   measured(4,[&]{parallel_jobs((take+8191)/8192,workers,[&](size_t job){for(size_t i=job*8192;i<std::min(take,(job+1)*8192);i++)a[i]=f.mul(a[i],b[i]);});});write(4*uint64_t(first),a.data(),take);first+=take;
  }};
  measured(3,[&]{zero(0,2*l);for(size_t first=0;first<k;){size_t take=std::min(tile,k-first);read(code_fd,p_offset+4*uint64_t(first),a.data(),take,uint64_t(1)<<f.s);if(!first&&a[0]!=1)throw std::runtime_error("disk P normalization");write(4*uint64_t(first),a.data(),take);first+=take;}read(code_fd,p_offset+4*uint64_t(k),b.data(),1,uint64_t(1)<<f.s);for(size_t first=0;first<k;){size_t take=std::min(tile,k-first);write(4*uint64_t(l+first),message+first,take);first+=take;}});
  fft(true,0,false);fft(true,4*uint64_t(l),false);product(true);fft(true,0,true);
  measured(3,[&]{zero(k,n-k);});fft(false,0,false);product(false);fft(false,0,true);
  const auto step=make_product(f,omega);uint32_t sum=0;
  for(size_t first=0;first<n-1;){size_t take=std::min(tile,n-1-first);read(data_fd,4*uint64_t(first),a.data(),take,f.q);
   measured(5,[&]{std::atomic<uint32_t> parity{0};std::atomic<bool> bad{false};constexpr size_t chunk=4096;parallel_jobs((take+chunk-1)/chunk,workers,[&](size_t job){uint32_t point=power(first+job*chunk+1),local=0;for(size_t i=job*chunk;i<std::min(take,(job+1)*chunk);i++){uint32_t value=a[i];if(value>=(1u<<f.s))bad.store(true);word[point]=value;local^=value;point=step.mul(f,point);}parity.fetch_xor(local);});if(bad.load())throw std::runtime_error("disk result outside base field");sum^=parity.load();});first+=take;
  }word[0]=sum;stats[13]=ns(begin);return 0;
 }catch(const std::exception& e){fail(error,capacity,e.what());return 1;}catch(...){fail(error,capacity,"disk online encoding failed");return 1;}
}
