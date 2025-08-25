#!/usr/bin/env python3
"""
BM42 인덱싱 검증 스크립트
인덱싱이 올바르게 되었는지 확인하고 통계를 제공합니다.

python3 scripts/verify_bm42_index.py --full --original json/gs_issues.json --report verification_report.json
"""

from qdrant_client import QdrantClient
from qdrant_client.models import Filter, FieldCondition, MatchValue
import json
import numpy as np
from collections import Counter
import argparse
import sys


class BM42IndexVerifier:
    def __init__(self, host="localhost", port=6333):
        """검증기 초기화"""
        self.client = QdrantClient(host=host, port=port)
        
    def verify_collection(self, collection_name):
        """컬렉션 기본 정보 확인"""
        try:
            info = self.client.get_collection(collection_name)
            
            print(f"\n📊 컬렉션 정보: {collection_name}")
            print("=" * 60)
            print(f"총 문서 수: {info.points_count:,}")
            print(f"상태: {info.status}")
            
            # Sparse Vector 설정 확인 (속성 존재 여부 체크)
            if hasattr(info.config, 'sparse_vectors_config') and info.config.sparse_vectors_config:
                print(f"\n🔧 Sparse Vector 설정:")
                for name, config in info.config.sparse_vectors_config.items():
                    if hasattr(config, 'modifier'):
                        print(f"  - {name}: IDF modifier = {config.modifier}")
                    else:
                        print(f"  - {name}: Sparse vector 설정됨")
                    
            return info.points_count
            
        except Exception as e:
            print(f"❌ 컬렉션 {collection_name}을 찾을 수 없습니다: {e}")
            return 0
            
    def analyze_sample_documents(self, collection_name, sample_size=10):
        """샘플 문서 분석"""
        print(f"\n🔍 샘플 문서 분석 (상위 {sample_size}개)")
        print("=" * 60)
        
        try:
            # 샘플 문서 가져오기
            samples = self.client.retrieve(
                collection_name=collection_name,
                ids=list(range(sample_size)),
                with_vectors=["bm42"],
                with_payload=True
            )
            
            token_counts = []
            value_stats = []
            
            for point in samples:
                print(f"\n📄 문서 ID: {point.id}")
                print(f"   Key: {point.payload.get('key', 'N/A')}")
                
                title = point.payload.get('title', '')
                if title:
                    print(f"   제목: {title[:80]}...")
                    
                if "bm42" in point.vector:
                    sparse = point.vector["bm42"]
                    num_tokens = len(sparse.indices)
                    token_counts.append(num_tokens)
                    
                    if sparse.values:
                        values = sparse.values
                        value_stats.extend(values)
                        
                        print(f"   토큰 수: {num_tokens}")
                        print(f"   최대 점수: {max(values):.4f}")
                        print(f"   평균 점수: {np.mean(values):.4f}")
                        print(f"   상위 5개 점수: {sorted(values, reverse=True)[:5]}")
                        
            # 전체 통계
            if token_counts:
                print(f"\n📈 전체 통계:")
                print(f"   평균 토큰 수: {np.mean(token_counts):.1f}")
                print(f"   토큰 수 범위: {min(token_counts)} ~ {max(token_counts)}")
                
            if value_stats:
                print(f"   전체 평균 점수: {np.mean(value_stats):.4f}")
                print(f"   점수 표준편차: {np.std(value_stats):.4f}")
                
        except Exception as e:
            print(f"❌ 샘플 분석 실패: {e}")
            
    def test_search_functionality(self, collection_name):
        """검색 기능 테스트"""
        print(f"\n🧪 검색 기능 테스트")
        print("=" * 60)
        
        test_queries = [
            ("EDR", "도메인 특화 용어"),
            ("보안", "한국어 일반 용어"),
            ("security", "영어 일반 용어"),
            ("Live Response", "복합 용어"),
            ("API 인증 오류", "한국어 + 영어 혼합"),
        ]
        
        from fastembed import SparseTextEmbedding
        model = SparseTextEmbedding("Qdrant/bm42-all-minilm-l6-v2-attentions")
        
        for query, description in test_queries:
            print(f"\n🔍 테스트: '{query}' ({description})")
            
            try:
                # 쿼리 임베딩 생성
                query_embedding = list(model.embed([query]))[0]
                
                from qdrant_client.models import SparseVector, NamedSparseVector
                
                query_vector = SparseVector(
                    indices=query_embedding.indices.tolist(),
                    values=query_embedding.values.tolist()
                )
                
                # 검색 실행
                results = self.client.search(
                    collection_name=collection_name,
                    query_vector=NamedSparseVector(
                        name="bm42",
                        vector=query_vector
                    ),
                    limit=3
                )
                
                print(f"   토큰 수: {len(query_embedding.indices)}")
                print(f"   결과 수: {len(results)}")
                
                for i, result in enumerate(results[:3], 1):
                    print(f"   {i}. [{result.payload.get('key')}] Score: {result.score:.4f}")
                    
            except Exception as e:
                print(f"   ❌ 검색 실패: {e}")
                
    def verify_data_integrity(self, collection_name, original_file=None):
        """데이터 무결성 검증"""
        print(f"\n✅ 데이터 무결성 검증")
        print("=" * 60)
        
        # Qdrant의 문서 수
        qdrant_count = self.client.get_collection(collection_name).points_count
        print(f"Qdrant 문서 수: {qdrant_count:,}")
        
        # 원본 파일과 비교
        if original_file:
            try:
                with open(original_file, 'r', encoding='utf-8') as f:
                    original_data = json.load(f)
                original_count = len(original_data)
                print(f"원본 파일 문서 수: {original_count:,}")
                
                if qdrant_count == original_count:
                    print("✅ 문서 수 일치!")
                else:
                    diff = abs(qdrant_count - original_count)
                    print(f"⚠️ 문서 수 불일치: {diff}개 차이")
                    
                # 샘플 키 확인
                if original_data and 'key' in original_data[0]:
                    sample_keys = [doc.get('key') for doc in original_data[:5] if 'key' in doc]
                    print(f"\n원본 샘플 키: {sample_keys}")
                    
                    # Qdrant에서 동일한 키 검색
                    for key in sample_keys:
                        filter_condition = Filter(
                            must=[
                                FieldCondition(
                                    key="key",
                                    match=MatchValue(value=key)
                                )
                            ]
                        )
                        
                        results = self.client.scroll(
                            collection_name=collection_name,
                            scroll_filter=filter_condition,
                            limit=1
                        )[0]
                        
                        if results:
                            print(f"  ✅ {key} 존재")
                        else:
                            print(f"  ❌ {key} 없음")
                            
            except Exception as e:
                print(f"❌ 원본 파일 읽기 실패: {e}")
                
    def generate_report(self, collection_name, output_file=None):
        """종합 보고서 생성"""
        report = {
            "collection": collection_name,
            "verification_results": {},
            "recommendations": []
        }
        
        # 기본 정보
        count = self.verify_collection(collection_name)
        report["verification_results"]["document_count"] = count
        
        if count == 0:
            report["recommendations"].append("컬렉션이 비어있습니다. 인덱싱을 다시 실행하세요.")
        elif count < 100:
            report["recommendations"].append("문서 수가 적습니다. 더 많은 데이터를 인덱싱하는 것을 고려하세요.")
            
        # 보고서 출력
        print(f"\n📋 검증 보고서")
        print("=" * 60)
        print(f"컬렉션: {collection_name}")
        print(f"문서 수: {count:,}")
        
        if report["recommendations"]:
            print(f"\n💡 권장사항:")
            for rec in report["recommendations"]:
                print(f"  - {rec}")
                
        # 파일로 저장
        if output_file:
            with open(output_file, 'w', encoding='utf-8') as f:
                json.dump(report, f, ensure_ascii=False, indent=2)
            print(f"\n보고서 저장: {output_file}")
            
        return report


def main():
    parser = argparse.ArgumentParser(description="BM42 인덱스 검증")
    parser.add_argument("--collection", "-c", default="jira_bm42_full", 
                        help="검증할 컬렉션 이름")
    parser.add_argument("--host", default="localhost", help="Qdrant 호스트")
    parser.add_argument("--port", type=int, default=6333, help="Qdrant 포트")
    parser.add_argument("--original", "-o", help="원본 JSON 파일 경로")
    parser.add_argument("--report", "-r", help="보고서 출력 파일")
    parser.add_argument("--full", action="store_true", help="전체 검증 실행")
    
    args = parser.parse_args()
    
    # 검증기 초기화
    verifier = BM42IndexVerifier(args.host, args.port)
    
    # 기본 검증
    count = verifier.verify_collection(args.collection)
    
    if count > 0:
        # 샘플 분석
        verifier.analyze_sample_documents(args.collection)
        
        if args.full:
            # 검색 테스트
            verifier.test_search_functionality(args.collection)
            
            # 데이터 무결성 검증
            if args.original:
                verifier.verify_data_integrity(args.collection, args.original)
                
        # 보고서 생성
        verifier.generate_report(args.collection, args.report)
        
    else:
        print(f"\n❌ 컬렉션 {args.collection}이 비어있거나 존재하지 않습니다.")
        sys.exit(1)


if __name__ == "__main__":
    main()