
import dredd_hooks as hooks


@hooks.before_all
def my_before_all_hook(transactions):
    print('before all')


@hooks.before_each
def my_before_each_hook(transaction):
    print('before each')


@hooks.before
def my_before_hook(transaction):
    print('before')


@hooks.before_each_validation
def my_before_each_validation_hook(transaction):
    print('before each validation')


@hooks.before_validation
def my_before_validation_hook(transaction):
    print('before validations')


@hooks.after
def my_after_hook(transaction):
    print('after')


@hooks.after_each
def my_after_each(transaction):
    print('after_each')


@hooks.after_all
def my_after_all_hook(transactions):
    print('after_all')
