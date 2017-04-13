package mongo

const (
	Unactived = 0
	Actived   = 1
)

type User struct {
	ID             string `bson:"_id" redis:"id"`
	Password       string `bson:"password" redis:"password"`
	Name           string `bson:"name" redis:"name"`
	Phone          string `bson:"phone" redis:"phone"`
	Email          string `bson:"email" redis:"email"`
	Group          string `bson:"group" redis:"group"`
	Index          int    `bson:"index" redis:"index"`
	Actived        int    `bson:"actived" redis:"actived"`
	Image          string `bson:"image" redis:"image"`
	Tag            string `bson:"tag" redis:"tag"`
	Chief          bool   `bson:"chief" redis:"chief"`
	Gender         int    `bson:"gender" redis:"gender"`
	Department     string `bson:"department" redis:"department"`
	UserType       int    `bson:"userType" redis:"userType"`
	ActiveCode     string `bson:"activeCode" redis:"activeCode"`
	ActiveTime     string `bson:"activeTime" redis:"activeTime"`
	LastLoginTime  string `bson:"lastLoginTime" redis:"-"`
	LastLoginAddr  string `bson:"lastLoginAddr" redis:"-"`
	JoinPartyDate  string `bson:"joinPartyDate" redis:"joinPartyDate"`
	BirthDay       string `bson:"birthDay" redis:"birthDay"`
	Title          string `bson:"title" redis:"title"` //职级
	Community      string `bson:"community" redis:"community"`
	Note           string `bson:"note" redis:"note"`                     //备注
	PopulationType string `bson:"populationType" redis:"populationType"` //人群类型
	TitleInParty   string `bson:"titleInParty" redis:"titleInParty"`     //党内职务
}

type UserArray []*User

func NewUser(id, phone, email, pwd, name, company, group, image, department, tag string, gender, index int, chief bool, userType int, activeCode, activeTime string) *User {
	return &User{
		ID:         id,
		Password:   pwd,
		Name:       name,
		Phone:      phone,
		Email:      email,
		Group:      group,
		Index:      index,
		Actived:    Unactived,
		Image:      image,
		Tag:        tag,
		Chief:      chief,
		Gender:     gender,
		Department: department,
		UserType:   userType,
		ActiveCode: activeCode,
		ActiveTime: activeTime,
		// Company:  company,
	}
}
