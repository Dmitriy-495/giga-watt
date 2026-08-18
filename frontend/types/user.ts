export interface User {
  id: number
  email: string
  is_superuser: boolean
}

export interface Employee {
  id: number
  first_name: string
  last_name: string
  middle_name?: string | null
}

export interface OrganizationUnit {
  id: number
  name: string
  type: string
  parent_id?: number | null
}

export interface OrganizationContext {
  current: OrganizationUnit | null
  ancestors: OrganizationUnit[]
}

export interface CurrentOperator {
  user: User
  employee: Employee | null
  organization_context: OrganizationContext | null
}
