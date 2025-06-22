# 호환성

이 API는 광범위한 기능과 사용 사례를 다루며 널리 구현되었다.
이러한 큰 기능 세트와 다양한 구현의 조합은
API가 사용되는 곳마다 일관된 경험을 제공하기 위해
명확한 호환성 정의와 테스트를 필요로 한다.

Gateway API 호환성을 고려할 때 세 가지 중요한 개념이 있다.

## 1. 릴리스 채널

Gateway API 내에서 릴리스 채널은 필드나 리소스의 안정성을 나타내는 데 사용된다.
API의 "standard" 채널에는 "beta"로 졸업한 필드와 리소스가 포함된다.
API의 "experimental" 채널에는 "standard" 채널의 모든 것과 함께
여전히 중대한 방식으로 변경되거나 **완전히 제거될 수 있는** 실험적인 필드와
리소스가 포함된다.
이 개념에 대한 자세한 내용은
[버전 관리](versioning.md) 문서를 참조하자.

## 2. 지원 수준

불행히도 API의 일부 구현은 정의된 모든 기능을 지원할 수 없을 것이다.
이를 해결하기 위해
API는 각 기능에 대한 해당 지원 수준을 정의한다.

* **Core** 기능은 이식 가능하며 이 범주의 API 지원을 위한 모든 구현에 대해
  합리적인 로드맵이 있을 것으로 기대한다.
* **Extended** 기능은 이식 가능하지만 구현에서 보편적으로 지원되지는
  않는 기능이다.
  해당 기능을 지원하는 구현은 동일한 동작과 의미를 갖게 된다.
  일부 로드맵 기능은 결국 Core로 마이그레이션될 것으로 예상된다.
  Extended 기능은 API 타입과 스키마의 일부가 될 것이다.
* **Implementation-specific** 기능은 이식 가능하지 않고 공급업체별 특정된 기능이다.
  Implementation-specific 기능은 일반적인 확장 지점을 통하지 않는 한
  API 타입과 스키마를 가지지 않을 것이다.

Core 및 Extended 세트의 동작과 기능은 행동 기반 호환성 테스트를 통해
정의되고 검증될 것이다. Implementation-specific 기능은 호환성 테스트에서
다루지 않는다.

API 사양에서 Extended 기능을 포함하고 표준화함으로써,
전체 API 지원을 손상시키지 않으면서 구현 간에 이식 가능한
API 하위 집합에 수렴할 수 있을 것으로 기대한다.
보편적 지원의 부족은 이식 가능한 기능 세트 개발에 걸림돌이 되지 않을 것이다.
사양에 대한 표준화는 지원이 광범위해질 때 결국 Core로 졸업하는 것을 더 쉽게 만들 것이다.

### 중첩된 지원 수준

특정 필드에 대해 지원 수준이 중첩될 수 있다.
이런 경우가 발생하면 표현된 최소 지원 수준으로 해석되어야 한다.
예를 들어, 동일한 구조체가 두 개의 다른 곳에 포함될 수 있다.
그 중 한 곳에서는 구조체가 Core 지원을 갖는 것으로 간주되고
다른 곳에서는 Extended 지원만 포함한다.
이 구조체 내의 필드는 별도의 Core 및 Extended 지원 수준을 나타낼 수 있지만,
이러한 수준은 포함된 상위 구조체의 지원 수준을 초과하는 것으로 해석되어서는 안 된다.

더 구체적인 예로, HTTPRoute는 Rule 내에서 정의된 필터에 대해 Core 지원을 포함하고,
BackendRef 내에서 정의된 경우 Extended 지원을 포함한다.
이러한 필터는 각 필드에 대해 별도로 지원 수준을 정의할 수 있다.
중첩된 지원 수준을 해석할 때 최소값으로 해석되어야 한다.
이는 필드가 Core 지원 수준을 가지더라도 Extended 지원을 갖는 곳에 연결된 필터에 있는 경우,
해석된 지원 수준은 Extended여야 함을 의미한다.

## 3. 호환성 테스트

게이트웨이 API는 호환성 테스트 세트를 포함한다.
이는 지정된 GatewayClass를 사용하여
일련의 Gateway와 Route를 생성하고, 구현이 API 사양과 일치하는지 테스트한다.

각 릴리스에는 호환성 테스트 세트가 포함되어 있으며,
API가 발전함에 따라 계속 확장될 것이다.
현재 호환성 테스트는 standard 채널의
Core 기능 대부분을 다루며, 일부 Extended 기능을 포함한다.

### 테스트 실행

호환성 테스트에는 두 가지 주요 대조적인 세트가 있다.

* 게이트웨이 관련 테스트 (인그레스 테스트로도 생각할 수 있음)
* 서비스 메시 관련 테스트

`Gateway` 테스트의 경우 `Gateway` 테스트 기능을 활성화한 후
실행하려는 특정 테스트(예: `HTTPRoute`)를 선택해야 한다.
메시 관련 테스트의 경우 `Mesh`를 활성화해야 한다.

각 사용 사례를 개별로 다루겠지만,
구현이 둘 다 지원하는 경우 이를 결합하는 것도 가능하다.
또한 실행 중인 테스트에 관계없이 전체 테스트 스위트에 적용되는 옵션도 있다.

#### 게이트웨이 테스트

기본적으로 `Gateway` 중심의 호환성 테스트는 클러스터에 `gateway-conformance`라는
이름의 GatewayClass가 설치되어 있을 것으로 예상하며, 이에 대해 테스트가
실행된다. 대부분의 경우 다른 클래스를 사용할 것이며, 이는 해당 테스트 명령과
함께 `-gateway-class` 플래그로 지정할 수 있다.
사용할 `gateway-class` 이름은 당신의 인스턴스에서 확인하자.
또한 `Gateway` 지원과 구현이 지원하는 모든 `*Routes`에
대한 테스트 지원을 활성화해야 한다.

다음은 `Gateway`, `HTTPRoute`, `ReferenceGrant`와 관련된 모든
테스트를 실행한다.

```shell
go test ./conformance -run TestConformance -args \
    --gateway-class=my-gateway-class \
    --supported-features=Gateway,HTTPRoute
```

다른 유용한 플래그는 [호환성 플래그][cflags]에서 찾을 수 있다.

[cflags]:https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/utils/flags/flags.go

#### 메시 테스트

Mesh 테스트는 단순히 `Mesh` 기능을 활성화하여 실행할 수 있다.

```shell
go test ./conformance -run TestConformance -args --supported-features=Mesh
```

메시가 `HTTPRoute`와 같은 API를 통해 인그레스 지원을 포함하고 있다면,
`Gateway` 기능과 관련 API 기능을 활성화하여 동일한 테스트 실행에서
관련 테스트를 실행할 수 있다.

```shell
go test ./conformance -run TestConformance -args --supported-features=Mesh,Gateway,HTTPRoute
```

#### 네임스페이스 레이블과 어노테이션

테스트에 사용되는 네임스페이스에 레이블이 필요한 경우, `-namespace-labels` 플래그를
사용하여 테스트 네임스페이스에 설정할 하나 이상의 `name=value` 레이블을 전달할 수 있다.
마찬가지로 `-namespace-annotations` 플래그를 사용하여 테스트 네임스페이스에 적용할
어노테이션을 지정할 수 있다.
메시 테스트의 경우, 이 플래그는 구현이 메시 워크로드를 호스팅하는 네임스페이스에 레이블을
요구하는 경우 사용할 수 있다. 예를 들어, 사이드카 주입을 화성화하기 위해 사용할 수 있다.

예를 들어, Linkerd를 테스트할 때 다음과 같이 실행할 수 있다.

```shell
go test ./conformance -run TestConformance -args \
   ...
   --namespace-annotations=linkerd.io/inject=enabled
```

이렇게 하면 테스트 네임스페이스가 메시에 올바르게 주입된다.

#### 테스트 제외

`Gateway` 및 `ReferenceGrant` 기능은 기본적으로 활성화된다.
`-supported-features` 플래그를 사용하여 명시적으로 나열할 필요는 없다.
그러나 실행하지 않으려면
`-exempt-features` 플래그를 사용하여 비활성화해야 한다.
예를 들어, `Mesh` 테스트만 실행하고 다른 것은 실행하지 않으려면 다음과 같다.

```shell
go test ./conformance -run TestConformance -args \
    --supported-features=Mesh \
    --exempt-features=Gateway,ReferenceGrant
```

#### 스위트 수준 옵션

어떤 종류의 테스트를 실행할 때 테스트 스위트가 완료 후
테스트 리소스를 정리하지 않도록 설정할 수 있다.
(즉, 실패 시 클러스터 상태를 검사할 수 있도록).
`--cleanup-base-resources=false` 플래그를 설정하여 정리를 건너뛸 수 있다.

(특히 특정 기능을 구현하는 작업을 할 때) 특정 테스트를 이름으로
실행하는 것이 도움이 될 수 있다.
이는 `--run-test` 플래그를 설정하여 수행할 수 있다.

#### 네트워크 정책

[Container Network Interface (CNI) 플러그인][network_plugins]을 사용하여
네트워크 정책을 시행하는 클러스터에서는
일부 호환성 테스트가 트래픽이 필요한 목적지에 도달할 수 있도록
사용자 정의 [`NetworkPolicy`][netpol] 리소스를 클러스터에 추가해야 할 수 있다.

사용자는 구현이 해당 테스트를 통과할 수 있도록
필요한 네트워크 정책을 추가해야 한다.

[network_plugins]: https://kubernetes.io/ko/docs/concepts/extend-kubernetes/compute-storage-net/network-plugins/
[netpol]: https://kubernetes.io/ko/docs/concepts/services-networking/network-policies/

### 호환성 프로필

호환성 프로필은 여러 지원 기능을 그룹화하여, 호환성 보고서를 통해
지원을 인증하는 것을 목표로 하는 도구이다.
지원되는 프로필은
`--conformance-profiles=PROFILE1,PROFILE2` 플래그를 사용하여 구성할 수 있다.

## 호환성 보고서

스위트가 호환성 프로필 세트를 실행하도록 구성되고 모든 호환성 보고서 필드가
[이 가이드][conformance-guide]에 따라 적절히 설정된 경우, 호환성 보고서는
PR을 통해 [이 폴더][reports-folder]에 업로드할 수 있다. 특정 구현에 대한
보고서를 구성하고 제출하는 방법에 대한 모든 세부 사항은 [가이드][conformance-guide]를
따르자.

[reports-folder]: https://github.com/kubernetes-sigs/gateway-api/tree/main/conformance/reports/
[conformance-guide]: https://github.com/kubernetes-sigs/gateway-api/tree/main/conformance/reports/README.md

## 호환성에 기여하기

많은 구현체는 전체 e2e 테스트 스위트의 일부로 호환성 테스트를 실행한다.
호환성 테스트에 기여한다는 것은 구현이 테스트 개발에 대한 투자를 공유하고
일관된 경험을 제공하고 있음을 보장할 수 있다는 것을
의미한다.

호환성과 관련된 모든 코드는 프로젝트의 "/conformance" 디렉터리에 있다.
테스트 정의는 "/conformance/tests"에 있으며 각 테스트는 두 개의 파일로 구성된다.
YAML 파일에는 테스트 실행 시 적용할 매니페스트가 포함되어 있다.
Go 파일에는 구현체가 해당 매니페스트를 적절히 처리하는지 확인하는 코드가
포함되어 있다.

호환성과 관련된 이슈는
["area/conformance"로 레이블이 지정](https://github.com/kubernetes-sigs/gateway-api/issues?q=is%3Aissue+is%3Aopen+label%3Aarea%2Fconformance)
되어 있다. 이는 종종 테스트 커버리지를 개선하기 위한 새로운 테스트 추가나 기존 테스트의
결함이나 제한 사항 수정을 다룬다.
