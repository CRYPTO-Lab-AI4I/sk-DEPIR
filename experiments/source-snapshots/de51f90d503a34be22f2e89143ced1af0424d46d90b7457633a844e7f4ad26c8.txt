// Included after matrix.cu primitives. Explicit host-backed transforms, no full
// device domain or retained host pointer; transform arithmetic stays on GPU.
// One thread owns a butterfly pair. Column batches interleave vectors; row
// batches are contiguous. Global row origins survive tiling of the domain.
struct BackedPair {size_t vector,group,j,length,vectors;};
__device__ BackedPair backed_pair(size_t i,size_t n,unsigned level,unsigned k,bool contiguous){
 size_t length=size_t(1)<<level,vectors=n/length,half=size_t(1)<<k;
 size_t vector=contiguous?i/(length/2):i%vectors,pair=contiguous?i%(length/2):i/vectors;
 return {vector,(pair/half)*2*half,pair%half,length,vectors};
}
__device__ size_t backed_index(const BackedPair& b,size_t coefficient,bool contiguous){return contiguous?b.vector*b.length+coefficient:coefficient*b.vectors+b.vector;}
__global__ void backed_quotient(uint32_t* data,size_t n,const Matrix* matrix,bool columns,bool contiguous,unsigned k,const uint32_t* log,const uint32_t* powers){
 size_t i=size_t(blockIdx.x)*blockDim.x+threadIdx.x;if(i>=n/2)return;
 const auto& m=*matrix;const auto& p=columns?m.columns:m.rows;unsigned level=columns?m.dimension-m.split:m.split;
 auto b=backed_pair(i,n,level,k,contiguous);size_t half=size_t(1)<<k;if(b.j>=half/2)return;
 size_t target=backed_index(b,b.group+half+b.j,contiguous);uint32_t value=data[target];
 // Lower degrees are <=half/2, so sources are in the unchanged high half.
 for(unsigned t=0;t+1<p.count[k];t++)if(p.degree[k][t]>b.j)value^=multiply(p.coef[k][t],data[backed_index(b,b.group+2*half+b.j-p.degree[k][t],contiguous)],m.s,m.delta,log,powers);
 data[target]=value;
}
__global__ void backed_butterfly(const uint32_t* src,uint32_t* dst,size_t n,const Matrix* matrix,bool columns,bool inverse,bool contiguous,unsigned k,size_t first_vector,const uint32_t* log,const uint32_t* powers){
 size_t i=size_t(blockIdx.x)*blockDim.x+threadIdx.x;if(i>=n/2)return;
 const auto& m=*matrix;const auto& p=columns?m.columns:m.rows;unsigned level=columns?m.dimension-m.split:m.split;
 auto b=backed_pair(i,n,level,k,contiguous);size_t half=size_t(1)<<k;
 uint32_t point=columns?0:uint32_t((b.vector+first_vector)*b.length);
 for(unsigned bit=0;bit<level;bit++)if((b.group>>bit)&1)point^=p.basis[bit];
 uint32_t c=0;for(unsigned bit=0;bit<2*m.s;bit++)if((point>>bit)&1)c^=p.values[k][bit];
 size_t lo=backed_index(b,b.group+b.j,contiguous),hi=backed_index(b,b.group+half+b.j,contiguous);uint32_t a=src[lo],q=src[hi];
 if(inverse){q=multiply(p.inverse[k],a^q,m.s,m.delta,log,powers);dst[lo]=a^multiply(c,q,m.s,m.delta,log,powers);dst[hi]=q;}
 else{
  for(unsigned t=0;t+1<p.count[k];t++)if(p.degree[k][t]<=b.j)a^=multiply(p.coef[k][t],src[backed_index(b,b.group+half+b.j-p.degree[k][t],contiguous)],m.s,m.delta,log,powers);
  uint32_t left=a^multiply(c,q,m.s,m.delta,log,powers);dst[lo]=left;dst[hi]=left^multiply(p.delta[k],q,m.s,m.delta,log,powers);
 }
}
__global__ void backed_unjoin(const uint32_t* src,uint32_t* dst,size_t n,const Matrix* matrix,bool columns,bool contiguous,unsigned k,const uint32_t* log,const uint32_t* powers){
 size_t i=size_t(blockIdx.x)*blockDim.x+threadIdx.x;if(i>=n/2)return;
 const auto& m=*matrix;const auto& p=columns?m.columns:m.rows;unsigned level=columns?m.dimension-m.split:m.split;
 auto b=backed_pair(i,n,level,k,contiguous);size_t half=size_t(1)<<k;
 size_t lo=backed_index(b,b.group+b.j,contiguous),hi=backed_index(b,b.group+half+b.j,contiguous);uint32_t a=src[lo],q=src[hi];
 // Read the immutable joined quotient for both remainder and quotient undo.
 for(unsigned t=0;t+1<p.count[k];t++){
  size_t d=p.degree[k][t];if(d<=b.j)a^=multiply(p.coef[k][t],src[backed_index(b,b.group+half+b.j-d,contiguous)],m.s,m.delta,log,powers);
  else q^=multiply(p.coef[k][t],src[backed_index(b,b.group+2*half+b.j-d,contiguous)],m.s,m.delta,log,powers);
 }dst[lo]=a;dst[hi]=q;
}
// Both arrays already belong to the bounded engine. Forward alternates them;
// inverse uses scratch for the joined quotient, then restores the input array.
static uint32_t* backed_tiles(Engine& e,const Matrix* device,const Matrix& m,size_t n,bool columns,bool inverse,size_t first_vector,double& seconds){
 auto* src=e.a.words();auto* dst=e.scratch.words();unsigned level=columns?m.dimension-m.split:m.split;bool contiguous=!columns;
 for(unsigned step=0;step<level;step++){
  unsigned k=inverse?step:level-1-step;
  if(!inverse){
   seconds+=measured([&]{backed_quotient<<<(n/2+255)/256,256>>>(src,n,device,columns,contiguous,k,e.log.words(),e.powers.words());});
   seconds+=measured([&]{backed_butterfly<<<(n/2+255)/256,256>>>(src,dst,n,device,columns,false,contiguous,k,first_vector,e.log.words(),e.powers.words());});std::swap(src,dst);
  }else{
   seconds+=measured([&]{backed_butterfly<<<(n/2+255)/256,256>>>(src,dst,n,device,columns,true,contiguous,k,first_vector,e.log.words(),e.powers.words());});
   seconds+=measured([&]{backed_unjoin<<<(n/2+255)/256,256>>>(dst,src,n,device,columns,contiguous,k,e.log.words(),e.powers.words());});
  }
 }return src;
}
extern "C" int gm_create_backed(unsigned tile,unsigned s,unsigned polynomial,unsigned delta,const uint32_t* logs,size_t lc,const uint32_t* powers,size_t pc,size_t max_points,size_t device_cap,void** out,char* error,size_t capacity){
 if(tile<4||tile>(1u<<26)||max_points<4||max_points>domain_limit||(max_points&(max_points-1))||device_cap<8*std::min(size_t(tile),max_points)+4*(lc+pc)+sizeof(Matrix)){if(error&&capacity)std::snprintf(error,capacity,"host-backed tile/domain/cap");return 1;}
 size_t small=4;while(small<=std::min(size_t(tile),max_points)/2)small*=2;
 int result=create_engine(false,true,false,false,s,polynomial,delta,logs,lc,powers,pc,small,device_cap,out,error,capacity);
 if(!result){auto& e=*static_cast<Engine*>(*out);e.host_backed=true;e.backed_tile=tile;e.max_points=max_points;}return result;
}
extern "C" int gm_apply_backed(void* handle,uint32_t* data,size_t n,int inverse,double* stats,char* error,size_t cap){return api(error,cap,[&]{
 if(!handle||!data||!stats||(inverse!=0&&inverse!=1))throw std::runtime_error("host-backed arguments");auto& e=*static_cast<Engine*>(handle);
 if(!e.host_backed||!e.host_plans.count(n))throw std::runtime_error("host-backed plan/state");e.grow(n);std::fill(stats,stats+12,0);auto begin=Clock::now();const auto& m=e.host_plans.at(n);auto* device=static_cast<Matrix*>(e.plans.at(n)->p);
 uint64_t q=uint64_t(1)<<(2*e.s);if(q<(uint64_t(1)<<32))for(size_t i=0;i<n;i++)if(data[i]>=q)throw std::runtime_error("host-backed data outside field");
 std::vector<uint32_t> stage(e.points);stats[10]=double(stage.capacity()*4);
 if(stage.capacity()>e.backed_tile)throw std::runtime_error("host staging bound");
 auto upload=[&](void* dst,const uint32_t* src,size_t count){auto t=Clock::now();check(cudaMemcpy(dst,src,count*4,cudaMemcpyHostToDevice));stats[0]+=elapsed(t);stats[8]+=double(count*4);};
 auto download=[&](uint32_t* dst,const void* src,size_t count){auto t=Clock::now();check(cudaMemcpy(dst,src,count*4,cudaMemcpyDeviceToHost));stats[5]+=elapsed(t);stats[9]+=double(count*4);};
 auto convert=[&]{for(unsigned step=0;step<m.dimension-m.split;step++){
  unsigned k=inverse?step:m.dimension-m.split-1-step;const auto& t=m.taylor[k];if(!t.count)continue;
  size_t half=size_t(1)<<(m.split+k),minimum=half,maximum=0;for(unsigned j=0;j<t.count;j++){minimum=std::min(minimum,size_t(t.degree[j]));maximum=std::max(maximum,size_t(t.degree[j]));}
  if(!minimum||maximum>half/2)throw std::runtime_error("host-backed Taylor dependency");
  auto pass=[&](bool quotient){for(size_t block=0;block<n;block+=2*half){size_t first=quotient?0:minimum,last=quotient?maximum:half;
   for(size_t j=first;j<last;){size_t count=std::min(e.points,last-j),target=block+(quotient?half:0)+j;upload(e.a.p,data+target,count);
    for(unsigned term=0;term<t.count;term++){
     size_t d=t.degree[term],lo=quotient?j:std::max(j,d),hi=quotient?std::min(j+count,d):j+count;if(lo>=hi)continue;
     size_t source=block+(quotient?2*half:half)+lo-d;upload(e.scratch.p,data+source,hi-lo);
     stats[1]+=measured([&]{backed_accumulate<<<(hi-lo+255)/256,256>>>(e.a.words(),e.scratch.words(),hi-lo,lo-j,t.coef[term],e.s,e.delta,e.log.words(),e.powers.words());});
    }download(data+target,e.a.p,count);j+=count;
   }
  }};
  // Quotient corrections read its unchanged high half. Remainder reads the
  // joined quotient in the disjoint upper half; inverse reverses these passes.
  if(inverse){pass(false);pass(true);}else{pass(true);pass(false);}
 }};
 size_t n1=size_t(1)<<m.split,n2=n/n1;
 auto columns=[&]{size_t batch=e.points/n2;for(size_t col=0;col<n1;){size_t count=std::min(batch,n1-col),words=count*n2;auto t=Clock::now();
  for(size_t row=0;row<n2;row++)std::copy(data+row*n1+col,data+row*n1+col+count,stage.data()+row*count);stats[6]+=elapsed(t);
  upload(e.a.p,stage.data(),words);auto* result=backed_tiles(e,device,m,words,true,inverse,0,stats[2]);download(stage.data(),result,words);t=Clock::now();
  for(size_t row=0;row<n2;row++)std::copy(stage.data()+row*count,stage.data()+(row+1)*count,data+row*n1+col);stats[6]+=elapsed(t);stats[11]++;col+=count;
 }};
 auto rows=[&]{size_t batch=e.points/n1;for(size_t row=0;row<n2;){size_t count=std::min(batch,n2-row),words=count*n1;upload(e.a.p,data+row*n1,words);
  auto* result=backed_tiles(e,device,m,words,false,inverse,row,stats[3]);download(data+row*n1,result,words);stats[11]++;row+=count;
 }};
 if(!inverse){convert();columns();rows();}else{rows();columns();convert();}stats[7]=elapsed(begin);
});}
// Generate canonical destinations on GPU. CPU scatters immutable source values;
// no N-sized orbit map is retained. A primitive generator makes them distinct.
__global__ void backed_points(uint32_t* points,size_t first,size_t count,uint32_t omega,unsigned s,uint32_t delta,const uint32_t* log,const uint32_t* powers){
 size_t i=size_t(blockIdx.x)*blockDim.x+threadIdx.x;if(i<count)points[i]=power(omega,first+i+1,s,delta,log,powers);
}
extern "C" int gm_encode_backed(void* handle,const uint32_t* p,size_t pc,const uint32_t* spectrum,const uint32_t* message,size_t k,uint32_t* arena,uint32_t* word,size_t n,uint32_t omega,double* stats,char* error,size_t cap){return api(error,cap,[&]{
 if(!handle||!p||!spectrum||!message||!arena||!word||!stats||arena==word)throw std::runtime_error("backed online arguments");auto& e=*static_cast<Engine*>(handle);
 if(!e.host_backed||n!=(size_t(1)<<(2*e.s))||n>e.max_points||!k||k>n/2||pc!=k+1||p[0]!=1)throw std::runtime_error("backed online shape/code");
 size_t l=4;while(l<2*k-1)l*=2;if(l>n/2||!e.plans.count(l)||!e.plans.count(n))throw std::runtime_error("backed online prefix/plans");
 uint64_t q=uint64_t(1)<<(2*e.s);for(size_t i=0;i<k;i++)if(message[i]>=(1u<<e.s))throw std::runtime_error("backed message outside base field");
 for(size_t i=0;i<pc;i++)if(p[i]>=(1u<<e.s))throw std::runtime_error("backed P outside base field");
 if(q<(uint64_t(1)<<32))for(size_t i=0;i<n;i++)if(spectrum[i]>=q)throw std::runtime_error("backed spectrum outside field");
 auto pow_host=[&](size_t exponent){return power(omega,exponent,e.s,e.delta,e.host_log.data(),e.host_powers.data());};
 if(!omega||uint64_t(omega)>=q||pow_host(n-1)!=1)throw std::runtime_error("backed generator invalid");
 size_t order=n-1;for(size_t d=2;d<=order/d;d++){if(order%d)continue;if(pow_host((n-1)/d)==1)throw std::runtime_error("backed generator nonprimitive");while(order%d==0)order/=d;}
 if(order>1&&pow_host((n-1)/order)==1)throw std::runtime_error("backed generator nonprimitive");
 std::fill(stats,stats+16,0);auto begin=Clock::now();stats[15]=double(4*n);
 auto upload=[&](void* dst,const uint32_t* src,size_t count){auto t=Clock::now();check(cudaMemcpy(dst,src,count*4,cudaMemcpyHostToDevice));stats[0]+=elapsed(t);stats[10]+=double(count*4);};
 auto download=[&](uint32_t* dst,const void* src,size_t count){auto t=Clock::now();check(cudaMemcpy(dst,src,count*4,cudaMemcpyDeviceToHost));stats[5]+=elapsed(t);stats[11]+=double(count*4);};
 auto transform=[&](uint32_t* data,size_t count,bool inverse){double st[12];char detail[512];if(gm_apply_backed(handle,data,count,inverse,st,detail,sizeof(detail)))throw std::runtime_error(detail);
  for(unsigned i=0;i<6;i++)stats[i]+=st[i];stats[13]+=st[6];stats[10]+=st[8];stats[11]+=st[9];stats[12]=std::max(stats[12],st[10]);stats[14]++;
 };
 auto product=[&](uint32_t* a,const uint32_t* b,size_t length){for(size_t first=0;first<length;){size_t count=std::min(e.points,length-first);upload(e.a.p,a+first,count);upload(e.scratch.p,b+first,count);
  stats[6]+=measured([&]{pointwise<<<(count+255)/256,256>>>(e.a.words(),e.scratch.words(),count,e.s,e.delta,e.log.words(),e.powers.words());});download(a+first,e.a.p,count);first+=count;
 }};
 auto timer=Clock::now();std::fill(arena,arena+2*l,0);std::copy(p,p+k,arena);std::copy(message,message+k,arena+l);stats[8]+=elapsed(timer);
 transform(arena,l,false);transform(arena+l,l,false);product(arena,arena+l,l);transform(arena,l,true);
 timer=Clock::now();std::fill(arena+k,arena+n,0);stats[8]+=elapsed(timer);
 transform(arena,n,false);product(arena,spectrum,n);transform(arena,n,true);
 std::vector<uint32_t> destinations(e.points);if(destinations.capacity()>e.backed_tile)throw std::runtime_error("backed canonical staging bound");stats[12]=std::max(stats[12],double(destinations.capacity()*4));uint32_t sum=0;
 for(size_t first=0;first<n-1;){size_t count=std::min(e.points,n-1-first);
  stats[9]+=measured([&]{backed_points<<<(count+255)/256,256>>>(e.a.words(),first,count,omega,e.s,e.delta,e.log.words(),e.powers.words());});download(destinations.data(),e.a.p,count);
  timer=Clock::now();for(size_t i=0;i<count;i++){uint32_t value=arena[first+i];if(value>=(1u<<e.s))throw std::runtime_error("backed result outside base field");word[destinations[i]]=value;sum^=value;}stats[13]+=elapsed(timer);first+=count;
 }word[0]=sum;stats[7]=elapsed(begin);
});}
