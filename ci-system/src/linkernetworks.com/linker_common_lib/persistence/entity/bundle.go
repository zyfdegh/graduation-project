package entity

type Bundle struct {
	ServiceGroupItem         ServiceGroup          `bson:"servicegroup" json:"servicegroup"`
	ConfigrationPackagesItem []AppContainerPackage `bson:"configrationpackages" json:"configrationpackages"`
}
