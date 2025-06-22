# GAMMA 이니셔티브 (서비스 메시를 위한 게이트웨이 API)

게이트웨이 API는 원래 클러스터 외부의 클라이언트에서 클러스터 내부의 서비스로의 트래픽을 관리하도록 설계되었다.
-- 즉, _인그레스_ 또는 [_북/남_][north/south traffic] 케이스 --.
하지만 시간이 지나면서 [서비스 메시] 사용자들의 관심에 따라
2022년에 GAMMA(**G**ateway **A**PI for **M**esh **M**anagement and **A**dministration) 이니셔티브가 만들어졌다.
이는 게이트웨이 API가 동일한 클러스터 내에서 서비스 간 또는
[_동/서_ 트래픽][east/west traffic]에도 사용될 수 있는 방법을
정의하기 위함이었다.

GAMMA 이니셔티브는 별도의 서브프로젝트가 아닌 게이트웨이 API 서브프로젝트 내의 전용 워크스트림으로,
[GAMMA 리드]가 관리한다.
GAMMA의 목표는 게이트웨이 API에 최소한의 변경만을 가하면서
서비스 메시를 구성하는 데 게이트웨이 API를 사용할 수 있는 방법을 정의하는 것이며,
항상 게이트웨이 API의 [역할 지향적] 특성을 보존한다.
또한 기술 스택이나 프록시에 관계없이 서비스 메시 프로젝트들의 게이트웨이 API 구현 간 일관성을 옹호하기 위해
노력한다.

## 결과물

GAMMA 이니셔티브의 작업은 메시 및 메시 관련 사용 사례를 다루기 위해 게이트웨이 API 명세를 확장하거나 개선하는
[게이트웨이 개선 제안서][geps]에 포함된다.
현재까지 이러한 변경사항들은 비교적 작았지만(때로는 상대적으로 큰 영향을 미치지만!),
앞으로도 그럴 것으로 예상한다.
게이트웨이 API 명세의 거버넌스는 게이트웨이 API 서브프로젝트의 메인테이너들에게만 남아
있다.

GAMMA 이니셔티브의 이상적인 최종 결과는
서비스 메시 사용 사례가 게이트웨이 API의 일급 관심사가 되는 것이며,
이 시점에는 더 이상 별도의 이니셔티브가 필요하지 않을 것이다.

## 기여

모든 수준의 기여자를 환영한다!
게이트웨이 API와 GAMMA에 [기여할 수 있는 다양한 방법][contributor-ladder]이 있으며
이는 기술적인 것과 비기술적인 것 모두를 포함한다.

시작하는 가장 간단한 방법은 정기적으로 열리는 게이트웨이 API [회의] 중 하나에
참석하는 것이다.

[north/south traffic]:../concepts/glossary.md#northsouth-traffic
[서비스 메시]:../concepts/glossary.md#service-mesh
[east/west traffic]:../concepts/glossary.md#eastwest-traffic
[역할 지향적]:../concepts/roles-and-personas.md
[geps]:../geps/overview.md
[contributor-ladder]:../contributing/contributor-ladder.md
[회의]:../contributing/index.md/#meetings
[GAMMA 리드]:https://github.com/kubernetes-sigs/gateway-api/blob/main/OWNERS_ALIASES#L23
