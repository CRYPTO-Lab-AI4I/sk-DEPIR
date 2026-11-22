#!/usr/bin/env python3
"""Check hierarchical group planning,inter-level bytes and input proofs."""
from pathlib import Path
import hashlib
import runpy
import tempfile
import unittest
HERE=Path(__file__).resolve().parent
H=runpy.run_path(str(HERE/'run.py'))


def encoded(coeff):return b''.join(v.to_bytes(4,'little')for v in coeff)


def product(a,b):
    out=[0]*(len(a)+len(b)-1)
    for i,x in enumerate(a):
        for j,y in enumerate(b):out[i+j]^=x&y
    return out


class ReviewTest(unittest.TestCase):
    def test_unequal_child_domains(self):
        plan=H['stages']([1025,1025,1023,1023])
        self.assertEqual([(r['First'],r['Ranks'],r['Domain'],r['Count'])for r in plan],[(0,2,4096,2049),(2,2,2048,2045),(0,4,4096,4093)])
        self.assertEqual([r['Domain']for r in plan[-1]['Sources']],[4096,2048])

    def fixture(self,d):
        rows=[dict(kind='host',rank=r,host=f'node{r//2}',visible_devices='0')for r in range(4)];label='roots-16';leaf=[1]*5
        def event(rank,kind,**kw):rows.append(dict(rank=rank,kind=kind,case=label,**kw))
        for rank in range(4):
            for suffix in ('leaf','leaf-reference'):(d/f'{label}-rank-{rank}-{suffix}.bin').write_bytes(encoded(leaf))
            event(rank,'root-tree',leaf_counts=[5]*4,levels=2,support_count=16,matches=True)
            event(rank,'root-reference',matches=True,witnesses=3,wide_indices=True)
        sources={(rank,1):leaf for rank in range(4)}
        fixture=runpy.run_path(str(HERE.parent/'distributed-gpu-product/test_review.py'))['ReviewTest']
        for stage in H['stages']([5]*4):
            first,size,n=stage['First'],stage['Ranks'],stage['Domain'];dim=n.bit_length()-1
            dest=d/f'{label}-level-{stage["Level"]}-group-{first}';dest.mkdir()
            sub=fixture().fixture(dest,cases=[(18,dim)],ranks=size)
            # Keep only first product shape;its operation sequence starts at1.
            sub=[r for r in sub if r.get('case')in(None,f's18-d{dim}-shape-0')]
            a,b=[sources[(first+s['First'],s['Ranks'])]for s in stage['Sources']];actual=product(a,b);padded=actual+[0]*(n-len(actual));sources[(first,size)]=actual
            for r in sub:
                if r['kind']=='host':continue
                rank=r['rank'];r.update(tree_case=label,tree_level=stage['Level'],group_first=first,group_size=size,group_rank=rank)
                if r['kind']=='completed':r['cases']=1
                if r['kind']=='measurement':
                    data=encoded(padded[rank*n//size:(rank+1)*n//size]);r.update(a_count=len(a),b_count=len(b),output_sha256=hashlib.sha256(data).hexdigest(),resident=True,right_spectrum_preserved=True,reused=True,local_taylor=True)
                    for suffix in ('output','expected'):(dest/f'rank-{rank}-s18-d{dim}-shape-0-{suffix}.bin').write_bytes(data)
                if r['kind']=='pointwise-check':r.update(words=min(n//size,17),checked_words=n//size,resident=True,alias_rejected=True,layout_rejected=True,invalid_rejected=True,third_preserved=True)
                r['rank']=first+rank;rows.append(r)
            for rank in range(size):
                rows.append(dict(rank=first+rank,kind='root-stage',tree_case=label,tree_level=stage['Level'],group_first=first,group_size=size,group_rank=rank,stage=stage,dimension=dim))
                event(first+rank,'root-group-closed',level=stage['Level'],first=first,matches=True)
            for operand,layout in enumerate(stage['Sources']):
                child=[a,b][operand];child_bytes=encoded(child+[0]*(layout['Domain']-len(child)))
                for source in range(layout['First'],layout['First']+layout['Ranks']):
                    lo_source=(source-layout['First'])*layout['Domain']//layout['Ranks'];hi_source=min(lo_source+layout['Domain']//layout['Ranks'],layout['Count'])
                    for target in range(size):
                        if source==target:continue
                        lo=max(lo_source,target*n//size);hi=min(hi_source,(target+1)*n//size)
                        if lo>=hi:continue
                        packet=dict(Operation=1,Domain=n,Operand=operand,Offset=lo,Words=hi-lo,Source=source,Target=target)
                        proof=hashlib.sha256(child_bytes[4*lo:4*hi]).digest().hex()*2
                        for rank,direction in ((source,'send'),(target,'receive')):
                            rows.append(dict(rank=first+rank,kind='operand-proof',tree_case=label,tree_level=stage['Level'],group_first=first,group_size=size,group_rank=rank,packet=packet,direction=direction,own_proof=proof,remote_proof=proof))
        final=sources[(0,4)];padded=final+[0]*(32-len(final))
        for rank in range(4):
            for suffix in ('final','reference'):(d/f'{label}-rank-{rank}-{suffix}.bin').write_bytes(encoded(padded[rank*8:(rank+1)*8]))
            event(rank,'root-final',domain=32,count=17,words=8,levels=2,matches=True,seconds=1)
            rows.append(dict(rank=rank,kind='hierarchy-completed',cases=1))
        return rows

    def test_complete_hierarchy_and_missing_interlevel_packet(self):
        with tempfile.TemporaryDirectory()as tmp:
            d=Path(tmp);rows=self.fixture(d)
            self.assertEqual(len(H['review'](rows,d,True)['measurements']),8)
            i=next(i for i,r in enumerate(rows)if r.get('kind')=='operand-proof'and r['tree_level']==2)
            with self.assertRaises(AssertionError):H['review'](rows[:i]+rows[i+1:],d,True)
            rows[i]['own_proof']='00'*64
            with self.assertRaises(AssertionError):H['review'](rows,d,True)

    def test_subgroups_not_repeated_global_products(self):
        old=H['VECTORS']['suite'];selected=[];H['VECTORS']['suite']=lambda c,**kw:selected.append(kw)
        try:
            H['suite'](dict(run=lambda *a:None,env={},go='go'))
            self.assertEqual(selected[0]['ranks'],4)
            self.assertEqual([a['test']for a in selected[0]['arms']],['TestHierarchicalRootMerge']*2)
            self.assertEqual([a['memcheck']for a in selected[0]['arms']],[False,True])
        finally:H['VECTORS']['suite']=old


if __name__=='__main__':unittest.main()
