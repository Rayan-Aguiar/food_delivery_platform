package application

import (
	"errors"

	apperrors "food_delivery_platform/shared/errors"
)

// mapNotFoundOrInternal traduz um erro de repositorio para o contrato de erro da
// aplicacao: sentinel vira 404, qualquer outro erro (infra) vira 500.
func mapNotFoundOrInternal(err error, sentinel error, notFoundMsg, internalMsg string) error {
	if errors.Is(err, sentinel) {
		return apperrors.NotFound(notFoundMsg, nil)
	}
	return apperrors.Internal(internalMsg, err)
}
