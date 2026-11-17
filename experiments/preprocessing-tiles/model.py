#!/usr/bin/env python3
"""Integer geometry and payload counts; no full-size timing or admission claim."""
import json


def padded(values):
    return max(4,1<<(values-1).bit_length())


def split_shape(n,k):
    if n<64 or n&(n-1) or not 0<k<n//2:
        raise ValueError('bounded two-by-seven model requires a power-of-two domain')
    block=n//8;p=k+1;g=n-k
    terms=[dict(a=min(block,p-i),b=min(block,g-j),offset=i+j)
           for i in range(0,p,block) for j in range(0,g,block)]
    if (p+block-1)//block!=2 or (g+block-1)//block!=7:
        raise ValueError('shape is not two-by-seven')
    for t in terms:
        t['result']=t['a']+t['b']-1;t['domain']=padded(t['result'])
        assert t['offset']+t['result']<=n
    return dict(n=n,k=k,block=block,p=p,g=g,subproducts=len(terms),
                max_subdomain=max(t['domain'] for t in terms),
                resident_device_arrays=24*n,
                split_device_arrays=24*max(t['domain'] for t in terms),
                retained_output_bytes=4*n,
                max_temporary_host_bytes=4*max(t['result'] for t in terms),
                resident_h2d_coefficient_bytes=4*(p+g),
                split_h2d_coefficient_bytes=4*sum(t['a']+t['b'] for t in terms),
                resident_d2h_coefficient_bytes=4*n,
                split_d2h_coefficient_bytes=4*sum(t['result'] for t in terms),
                transform_work_ratio=sum(t['domain']*(t['domain'].bit_length()-1) for t in terms)/(n*(n.bit_length()-1)))


def check():
    for power in (6,9,22,36):
        n=1<<power;k=n*16300262095//(1<<36)
        m=split_shape(n,k)
        assert m['subproducts']==14 and m['max_subdomain']==n//4
        assert m['resident_device_arrays']==4*m['split_device_arrays']
    full=split_shape(1<<36,16300262095)
    assert full['resident_device_arrays']==1536<<30
    assert full['split_device_arrays']==384<<30
    assert full['max_temporary_host_bytes']==(64<<30)-4
    return full


if __name__=='__main__':
    print(json.dumps({'s18_projection':check(),'bounded_component':split_shape(1<<22,(1<<22)*16300262095//(1<<36)),
                      'scope':'Payload model excludes runtime/context/cache; work ratio is not a measured time penalty.'},indent=2))
