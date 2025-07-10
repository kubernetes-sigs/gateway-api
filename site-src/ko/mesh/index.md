
# 서비스 메시를 위한 게이트웨이 API

??? success "v1.1.0부터 표준 채널"

    서비스 메시 사용 사례를 지원하기 위한 [GAMMA 이니셔티브](gamma.md) 작업은
    v1.1.0부터 표준 채널에 포함되었으며 GA로 간주된다.
    자세한 정보는 [버전 관리 가이드](../concepts/versioning.md)를 참조하라.

"[GAMMA 이니셔티브](gamma.md)"는 게이트웨이 API가 서비스 메시에서
어떻게 사용될 수 있는지를 정의하는 그룹을 의미한다.
현재까지 이 그룹은 비교적 작은 변경사항으로 게이트웨이 API에서 서비스 메시 지원을 정의할 수 있었다.
GAMMA가 현재까지 도입한 가장 중요한 변경사항은 서비스 메시를 구성할 때
개별 라우트 리소스([HTTPRoute] 등)가 [서비스 리소스와 직접 연결](#gateway-api-for-mesh)된다는
것이다.

이는 주로 클러스터에서 일반적으로 하나의 메시만 활성화되므로
메시 작업 시 [게이트웨이]와 [게이트웨이 클래스] 리소스가 사용되지 않기 때문이다.
결과적으로 서비스 리소스가 라우팅 정보를 위한 가장 범용적인 바인딩 지점으로 남게
된다.

서비스 리소스는 불행히도 복잡하며 여러 오버로드되거나 명세가 부족한 측면이 있기 때문에,
GAMMA는 [서비스 _프론트엔드_ 와 _백엔드_ 측면][service-facets]을 공식적으로 정의하는 것이 중요하다고 판단했다.
간단히 말하면,

- 서비스 프론트엔드는 그 이름과 클러스터 IP이고,
- 서비스 백엔드는 엔드포인트 IP들의 집합이다.

이러한 구분은 게이트웨이 API가 서비스를 크게 복제하는 새로운 리소스를 요구하지 않으면서도
메시 내에서 라우팅이 어떻게 작동하는지 정확하게 설명하는 데
도움이 된다.

[게이트웨이 클래스]: ../api-types/gatewayclass.md
[게이트웨이]: ../api-types/gateway.md
[HTTPRoute]: ../api-types/httproute.md
[TCPRoute]: ../concepts/api-overview.md#tcproute-and-udproute
[Service]: https://kubernetes.io/docs/concepts/services-networking/service/
[service-mesh]:../concepts/glossary.md#service-mesh
[service-facets]:service-facets.md

## 라우트와 서비스 연결 <a name="gateway-api-for-mesh">

GAMMA는 개별 라우트 리소스가 서비스에 직접 연결되어
_해당 서비스로 향하는 모든 트래픽에 적용될 구성_ 을
나타낸다고 명시한다.

하나 이상의 라우트가 서비스에 연결되면
**최소 하나 이상의 라우트와 일치하지 않는 요청은 거부된다**.
라우트가 서비스에 연결되지 않은 경우,
해당 서비스에 대한 요청은 단순히 메시의 기본 동작에 따라 진행된다
(일반적으로 메시가 없는 것처럼 요청이 전달됨).

주어진 서비스에 어떤 라우트가 연결되는지는
라우트 자체에서 제어한다(쿠버네티스 RBAC와 함께 작동).
라우트는 게이트웨이가 아닌 서비스를 지정하는 `parentRef`를 단순히 명시한다.

```yaml
kind: HTTPRoute
metadata:
  name: smiley-route
  namespace: faces
spec:
  parentRefs:
    - name: smiley
      kind: Service
      group: core
      port: 80
  rules:
    ...
```

!!! note "진행 중인 작업"

    프로듀서 라우트와 컨슈머 라우트 간의 관계에 대한 작업이
    진행 중이다.

라우트의 네임스페이스와 서비스의 네임스페이스 간의 관계가
중요하다.

- 동일한 네임스페이스 <a name="producer-routes"></a>

    서비스와 동일한 네임스페이스에 있는 라우트를 [프로듀서 라우트]라고 하는데, 
    이는 일반적으로 워크로드의 생성자가 워크로드의 허용 가능한 사용을 정의하기 위해생성하기 때문이다
    (예를 들어, [Ana]가 워크로드와 라우트를 모두 배포하는 경우).
    모든 네임스페이스의 워크로드 클라이언트로부터의 모든 요청이
    이 라우트의 영향을 받는다.

    위에 표시된 라우트는 프로듀서 라우트이다.

- 다른 네임스페이스 <a name="consumer-routes"></a>

    서비스와 다른 네임스페이스에 있는 라우트를 [컨슈머 라우트]라고 한다.
    일반적으로 이는 주어진 워크로드의 소비자가 해당 워크로드에 대한 요청 방식을 개선하기 위한 라우트이다
    (예를 들어, 해당 소비자의 워크로드 사용에 대한 사용자 정의 타임아웃 구성).
    이 라우트는 라우트와 동일한 네임스페이스의 워크로드로부터의 요청에만
    영향을 준다.

    예를 들어, 아래의 HTTPRoute는 `fast-clients` 네임스페이스에 있는 `smiley` 워크로드의 모든 클라이언트가
    100ms 타임아웃을 갖도록 한다.

    ```yaml
    kind: HTTPRoute
    metadata:
      name: smiley-route
      namespace: fast-clients
    spec:
      parentRefs:
      - name: smiley
        namespace: faces
        kind: Service
        group: core
        port: 80
      rules:
        ...
        timeouts:
          request: 100ms
    ```

서비스에 바인딩된 라우트에 대해 중요한 참고 사항은
단일 네임스페이스 내에서 동일한 서비스에 대한 여러 라우트(프로듀서 라우트든 컨슈머 라우트든)가
게이트웨이 API [라우트 병합 규칙]에 따라 결합된다는 것이다.
따라서 현재 동일한 네임스페이스 내의 여러 소비자에 대해 구별되는 컨슈머 라우트를 정의하는 것은
불가능하다.

예를 들어, `blender` 워크로드와 `mixer` 워크로드가 모두 `foodprep` 네임스페이스에 있고,
둘 다 동일한 서비스를 사용하여 `oven` 워크로드를 호출하는 경우,
현재 `blender`와 `mixer`가 HTTPRoute를 사용하여 `oven` 워크로드 호출에 대해
서로 다른 타임아웃을 설정하는 것은 불가능하다.
이를 허용하려면 `blender`와 `mixer`를 별도의 네임스페이스로 이동해야 한다.

[Ana]:../concepts/roles-and-personas.md#ana
[프로듀서 라우트]:../concepts/glossary.md#producer-route
[컨슈머 라우트]:../concepts/glossary.md#consumer-route
[서비스 메시]:../concepts/glossary.md#service-mesh
[라우트 병합 규칙]:../api-types/httproute.md#merging

## 요청 흐름

GAMMA 호환 메시가 사용될 때 일반적인 [동/서] API 요청 흐름은 다음과
같다.

1. 클라이언트 워크로드가 <http://foo.ns.service.cluster.local>로 요청을 한다.
2. 메시 데이터 플레인이 요청을 가로채고 이를 네임스페이스 `ns`의 서비스 `foo`에 대한 트래픽으로
   식별한다.
3. 데이터 플레인이 `foo` 서비스와 연결된 라우트를 찾은 다음,

    a. 연결된 라우트가 없으면 요청이 항상 허용되고,
       `foo` 워크로드 자체가 대상 워크로드로 간주된다.

    b. 연결된 라우트가 있고 요청이 그 중 최소 하나와 일치하면,
       가장 높은 우선순위의 일치하는 라우트의 `backendRefs`가 대상 워크로드를 선택하는 데
       사용된다.

    c. 연결된 라우트가 있지만 요청이 그 중 어느 것과도 일치하지 않으면 요청이
       거부된다.

6. 데이터 플레인이 대상 워크로드로 요청을 라우팅한다
   (대부분 [엔드포인트 라우팅]을 사용하지만 [서비스 라우팅]을 사용할 수도 있다).

[동/서]:../concepts/glossary.md#eastwest-traffic
[엔드포인트 라우팅]:../concepts/glossary.md#endpoint-routing
[서비스 라우팅]:../concepts/glossary.md#service-routing
